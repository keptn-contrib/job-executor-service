package k8sutils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"keptn-contrib/job-executor-service/pkg/utils"

	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnutils "github.com/keptn/kubernetes-utils/pkg"
	"k8s.io/client-go/kubernetes"

	"keptn-contrib/job-executor-service/pkg/config"

	"github.com/PaesslerAG/jsonpath"
	"gopkg.in/yaml.v2"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const envValueFromEvent = "event"
const envValueFromSecret = "secret"
const envValueFromString = "string"

const minTTLSecondsAfterFinished = int32(60)
const defaultTTLSecondsAfterFinished = int32(21600)

const reasonJobDeadlineExceeded = "DeadlineExceeded"

// ErrPrivilegedContainerNotAllowed indicates an error that occurs if a security context does contain privileged=true
// but the policy of the job-executor-service doesn't allow such job workloads to be created
var /*const*/ ErrPrivilegedContainerNotAllowed = errors.New("privileged containers are not allowed")

// ErrMaxPollTimeExceeded indicates that the job has been polled for max poll time without completing.
// K8sImpl will stop polling abd return the error but the job will continue running on K8s.
var /*const*/ ErrMaxPollTimeExceeded = errors.New("max poll count reached for job")

// ErrTaskDeadlineExceeded indicates that the job has exceeded the deadline set for task runs.
// K8s has terminated the job and the related pods.
var /*const*/ ErrTaskDeadlineExceeded = errors.New("job deadline exceeded")

// JobSettings contains environment variable settings for the job
type JobSettings struct {
	JobNamespace                string
	KeptnAPIToken               string
	InitContainerImage          string
	DefaultResourceRequirements *v1.ResourceRequirements
	DefaultJobServiceAccount    string
	DefaultSecurityContext      *v1.SecurityContext
	DefaultPodSecurityContext   *v1.PodSecurityContext
	AllowPrivilegedJobs         bool
	TaskDeadlineSeconds         *int64
}

// K8sImpl is used to interact with kubernetes jobs
type K8sImpl struct {
	clientset kubernetes.Interface
}

// NewK8s creates and returns new K8s
func NewK8s(namespace string) *K8sImpl {
	return &K8sImpl{}
}

// ConnectToCluster returns the k8s Clientset
func (k8s *K8sImpl) ConnectToCluster() error {

	config, err := keptnutils.GetClientset(true)
	if err != nil {
		return err
	}
	k8s.clientset = config

	return nil
}

// CreateK8sJob creates a k8s job with the job-executor-service-initcontainer and the job image of the task
// returns ttlSecondsAfterFinished as stored in k8s or error in case of issues creating the job
func (k8s *K8sImpl) CreateK8sJob(
	jobName string, action *config.Action, task config.Task, eventData keptn.EventProperties, jobSettings JobSettings,
	jsonEventData interface{}, namespace string,
) error {

	var backOffLimit int32 = 0

	jobVolumeName := "job-volume"

	// TODO configure from outside:
	jobVolumeMountPath := "/keptn"

	// TODO configure from outside:
	quantity := resource.MustParse("20Mi")

	jobResourceRequirements := jobSettings.DefaultResourceRequirements
	if task.Resources != nil {
		var err error
		jobResourceRequirements, err = CreateResourceRequirements(
			task.Resources.Limits.CPU,
			task.Resources.Limits.Memory,
			task.Resources.Requests.CPU,
			task.Resources.Requests.Memory,
		)
		if err != nil {
			return fmt.Errorf("unable to create resource requirements for task %v: %v", task.Name, err.Error())
		}
	}

	emptyDirVolume := v1.EmptyDirVolumeSource{
		Medium:    v1.StorageMediumDefault,
		SizeLimit: &quantity,
	}

	// Use default service account but allow overriding
	// from the task configuration
	serviceAccountName := jobSettings.DefaultJobServiceAccount
	if task.ServiceAccount != nil {
		serviceAccountName = *task.ServiceAccount
	}

	jobEnv, err := k8s.prepareJobEnv(task, eventData, jsonEventData, namespace)
	if err != nil {
		return fmt.Errorf("could not prepare env for job %v: %v", jobName, err.Error())
	}

	// Set the default value of the ttlSecondsAfterFinished to ensure that the jobs are cleanup after some
	// time, if the TTL is too low it will be set to minTTLSecondsAfterFinished and a warning will be printed
	TTLSecondsAfterFinished := defaultTTLSecondsAfterFinished
	if task.TTLSecondsAfterFinished != nil {

		if *task.TTLSecondsAfterFinished < minTTLSecondsAfterFinished {
			TTLSecondsAfterFinished = minTTLSecondsAfterFinished
			log.Printf("Warning: Correcting TTLSecondsAfterFinished in action '%s' for task '%s' to %d!",
				action.Name, task.Name, minTTLSecondsAfterFinished,
			)
		} else {
			TTLSecondsAfterFinished = *task.TTLSecondsAfterFinished
		}

	}

	// Build the final security context for the pod
	jobSecurityContext := utils.BuildSecurityContext(jobSettings.DefaultSecurityContext, task.SecurityContext)

	// Warn the user if the resulting security context does contain any bad properties
	violations := utils.CheckJobSecurityContext(jobSecurityContext)
	if len(violations) != 0 {
		log.Printf("WARNING: Job %v has a potential insecure job securityContext!", jobName)
	}

	// If the privileged flag is contained check if these type of workloads are allowed and
	// abort the execution if they aren't or warn the user that such jobs are a bad idea
	if jobSecurityContext.Privileged != nil && *jobSecurityContext.Privileged {
		if jobSettings.AllowPrivilegedJobs {
			log.Printf("WARNING: Job %s will be executed in a privileged container", jobName)
		} else {
			return ErrPrivilegedContainerNotAllowed
		}
	}

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{

					SecurityContext: jobSettings.DefaultPodSecurityContext,
					InitContainers: []v1.Container{
						{
							Name:            "init-" + jobName,
							Image:           jobSettings.InitContainerImage,
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: jobSecurityContext,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      jobVolumeName,
									MountPath: jobVolumeMountPath,
								},
							},
							Env: []v1.EnvVar{
								{
									Name: "CONFIGURATION_SERVICE",
									ValueFrom: &v1.EnvVarSource{
										ConfigMapKeyRef: &v1.ConfigMapKeySelector{
											LocalObjectReference: v1.LocalObjectReference{
												Name: "job-service-config",
											},
											Key: "init_container_configuration_endpoint",
										},
									},
								},
								{
									Name: "KEPTN_API_TOKEN",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{
												Name: "job-service-keptn-secrets",
											},
											Key: "token",
										},
									},
								},
								{
									Name:  "KEPTN_PROJECT",
									Value: eventData.GetProject(),
								},
								{
									Name:  "KEPTN_STAGE",
									Value: eventData.GetStage(),
								},
								{
									Name:  "KEPTN_SERVICE",
									Value: eventData.GetService(),
								},
								{
									Name:  "JOB_ACTION",
									Value: action.Name,
								},
								{
									Name:  "JOB_TASK",
									Value: task.Name,
								},
							},
							Resources: *jobSettings.DefaultResourceRequirements,
						},
					},
					Containers: []v1.Container{
						{
							Name:            jobName,
							Image:           task.Image,
							ImagePullPolicy: v1.PullPolicy(task.ImagePullPolicy),
							Command:         task.Cmd,
							Args:            task.Args,
							WorkingDir:      task.WorkingDir,
							SecurityContext: jobSecurityContext,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      jobVolumeName,
									MountPath: jobVolumeMountPath,
								},
							},
							Env:       jobEnv,
							Resources: *jobResourceRequirements,
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
					Volumes: []v1.Volume{
						{
							Name: jobVolumeName,
							VolumeSource: v1.VolumeSource{
								EmptyDir: &emptyDirVolume,
							},
						},
					},
					ServiceAccountName: serviceAccountName,
				},
			},
			BackoffLimit:            &backOffLimit,
			TTLSecondsAfterFinished: &TTLSecondsAfterFinished,
			ActiveDeadlineSeconds:   jobSettings.TaskDeadlineSeconds,
		},
	}

	jobs := k8s.clientset.BatchV1().Jobs(namespace)

	_, err = jobs.Create(context.TODO(), jobSpec, metav1.CreateOptions{})

	if err != nil {
		return err
	}

	return nil
}

// AwaitK8sJobDone will poll the job status every pollInterval up to maxPollDuration.
// If the job completes successfully before we reach maxPollDuration, no error is returned.
// If the job fails, is suspended or does not complete within maxPollDuration, an appropriate error will be returned
func (k8s *K8sImpl) AwaitK8sJobDone(
	jobName string, maxPollDuration time.Duration, pollInterval time.Duration, namespace string,
) error {
	jobs := k8s.clientset.BatchV1().Jobs(namespace)

	pollingStart := time.Now()

	for {

		now := time.Now()

		if now.After(pollingStart.Add(maxPollDuration)) {
			return fmt.Errorf(
				"polling for job %s timing out after %s: %w", jobName, now.Sub(pollingStart),
				ErrMaxPollTimeExceeded,
			)
		}

		job, err := jobs.Get(context.TODO(), jobName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		for _, condition := range job.Status.Conditions {

			switch condition.Type {
			case batchv1.JobComplete:
				// hooray, it worked
				return nil
			case batchv1.JobSuspended:
				return fmt.Errorf(
					"job %s was suspended. Reason: %s, Message: %s", jobName, condition.Reason, condition.Message,
				)
			case batchv1.JobFailed:
				if condition.Reason == reasonJobDeadlineExceeded {
					return fmt.Errorf("job %s failed: %w", jobName, ErrTaskDeadlineExceeded)
				}

				return fmt.Errorf(
					"job %s failed. Reason: %s, Message: %s", jobName, condition.Reason, condition.Message,
				)
			}
		}

		time.Sleep(pollInterval)
	}
}

func (k8s *K8sImpl) prepareJobEnv(
	task config.Task, eventData keptn.EventProperties, jsonEventData interface{}, namespace string,
) ([]v1.EnvVar, error) {

	var jobEnv []v1.EnvVar
	for _, env := range task.Env {
		var err error
		var generatedEnv []v1.EnvVar

		switch env.ValueFrom {
		case envValueFromEvent:
			generatedEnv, err = generateEnvFromEvent(env, jsonEventData)
		case envValueFromSecret:
			generatedEnv, err = k8s.generateEnvFromSecret(env, namespace)
		case envValueFromString:
			generatedEnv = generateEnvFromString(env)
		default:
			return nil, fmt.Errorf("could not add env with name %v, unknown valueFrom %v", env.Name, env.ValueFrom)
		}

		if err != nil {
			return nil, err
		}
		jobEnv = append(jobEnv, generatedEnv...)
	}

	// append KEPTN_PROJECT, KEPTN_SERVICE and KEPTN_STAGE as environment variables
	jobEnv = append(
		jobEnv,
		v1.EnvVar{
			Name:  "KEPTN_PROJECT",
			Value: eventData.GetProject(),
		},
		v1.EnvVar{
			Name:  "KEPTN_STAGE",
			Value: eventData.GetStage(),
		},
		v1.EnvVar{
			Name:  "KEPTN_SERVICE",
			Value: eventData.GetService(),
		},
	)

	replacer := strings.NewReplacer("-", "_", " ", "_")

	// append labels as environment variables
	for key, value := range eventData.GetLabels() {
		// replace - with _
		key = replacer.Replace(key)

		jobEnv = append(
			jobEnv,
			v1.EnvVar{
				Name:  "LABELS_" + strings.ToUpper(key),
				Value: value,
			},
		)
	}

	return jobEnv, nil
}

func generateEnvFromEvent(env config.Env, jsonEventData interface{}) ([]v1.EnvVar, error) {

	value, err := jsonpath.Get(env.Value, jsonEventData)
	if err != nil {
		return nil, fmt.Errorf(
			"could not add env with name '%v', value '%v', valueFrom '%v': %v", env.Name, env.Value, env.ValueFrom, err,
		)
	}

	if strings.EqualFold(env.Formatting, "yaml") {
		yamlString, err := yaml.Marshal(value)

		if err != nil {
			return nil, fmt.Errorf(
				"could not convert env with name '%v', value '%v', valueFrom '%v' to YAML: %v", env.Name, env.Value,
				env.ValueFrom, err,
			)
		}

		value = string(yamlString[:])
	} else if strings.EqualFold(env.Formatting, "json") || reflect.ValueOf(value).Kind() == reflect.Map {
		jsonString, err := json.Marshal(value)

		if err != nil {
			return nil, fmt.Errorf(
				"could not convert env with name '%v', value '%v', valueFrom '%v' to JSON: %v", env.Name, env.Value,
				env.ValueFrom, err,
			)
		}

		value = string(jsonString[:])
	}

	generatedEnv := []v1.EnvVar{
		{
			Name:  env.Name,
			Value: fmt.Sprintf("%v", value),
		},
	}

	return generatedEnv, nil
}

func generateEnvFromString(env config.Env) []v1.EnvVar {
	return []v1.EnvVar{
		{
			Name:  env.Name,
			Value: env.Value,
		},
	}
}

func (k8s *K8sImpl) generateEnvFromSecret(env config.Env, namespace string) ([]v1.EnvVar, error) {

	var generatedEnv []v1.EnvVar

	secret, err := k8s.clientset.CoreV1().Secrets(namespace).Get(context.TODO(), env.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not add env with name %v, valueFrom %v: %v", env.Name, env.ValueFrom, err)
	}

	for key := range secret.Data {
		generatedEnv = append(
			generatedEnv, v1.EnvVar{
				Name: key,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{Name: env.Name},
						Key:                  key,
					},
				},
			},
		)
	}

	return generatedEnv, nil
}
