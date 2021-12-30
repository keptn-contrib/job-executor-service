package k8sutils

import (
	"context"
	"encoding/json"
	"fmt"
	"keptn-contrib/job-executor-service/pkg/config"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"gopkg.in/yaml.v2"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const envValueFromEvent = "event"
const envValueFromSecret = "secret"
const envValueFromString = "string"

// JobSettings contains environment variable settings for the job
type JobSettings struct {
	JobNamespace                                 string
	InitContainerConfigurationServiceAPIEndpoint string
	KeptnAPIToken                                string
	InitContainerImage                           string
	DefaultResourceRequirements                  *v1.ResourceRequirements
	AlwaysSendFinishedEvent                      bool
	EnableKubernetesAPIAccess                    bool
}

// CreateK8sJob creates a k8s job with the job-executor-service-initcontainer and the job image of the task
func (k8s *k8sImpl) CreateK8sJob(jobName string, action *config.Action, task config.Task, eventData *keptnv2.EventData,
	jobSettings JobSettings, jsonEventData interface{}, namespace string) error {

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
	automountServiceAccountToken := jobSettings.EnableKubernetesAPIAccess

	// specify empty service account name for job
	serviceAccountName := ""

	if jobSettings.EnableKubernetesAPIAccess {
		automountServiceAccountToken = true
		serviceAccountName = "job-executor-service"
	}

	runAsNonRoot := true
	convert := func(s int64) *int64 {
		return &s
	}

	jobEnv, err := k8s.prepareJobEnv(task, eventData, jsonEventData, namespace)
	if err != nil {
		return fmt.Errorf("could not prepare env for job %v: %v", jobName, err.Error())
	}

	var TTLSecondsAfterFinished int32
	if task.TTLSecondsAfterFinished == nil {
		TTLSecondsAfterFinished = 21600
	} else {
		TTLSecondsAfterFinished = *task.TTLSecondsAfterFinished
	}

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{

					SecurityContext: &v1.PodSecurityContext{
						RunAsUser:    convert(1000),
						RunAsGroup:   convert(2000),
						FSGroup:      convert(2000),
						RunAsNonRoot: &runAsNonRoot,
					},
					InitContainers: []v1.Container{
						{
							Name:            "init-" + jobName,
							Image:           jobSettings.InitContainerImage,
							ImagePullPolicy: v1.PullAlways,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      jobVolumeName,
									MountPath: jobVolumeMountPath,
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "CONFIGURATION_SERVICE",
									Value: jobSettings.InitContainerConfigurationServiceAPIEndpoint,
								},
								{
									Name:  "KEPTN_API_TOKEN",
									Value: jobSettings.KeptnAPIToken,
								},
								{
									Name:  "KEPTN_PROJECT",
									Value: eventData.Project,
								},
								{
									Name:  "KEPTN_STAGE",
									Value: eventData.Stage,
								},
								{
									Name:  "KEPTN_SERVICE",
									Value: eventData.Service,
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
							Name:       jobName,
							Image:      task.Image,
							Command:    task.Cmd,
							Args:       task.Args,
							WorkingDir: task.WorkingDir,
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
					AutomountServiceAccountToken: &automountServiceAccountToken,
					ServiceAccountName:           serviceAccountName,
				},
			},
			BackoffLimit:            &backOffLimit,
			TTLSecondsAfterFinished: &TTLSecondsAfterFinished,
		},
	}

	jobs := k8s.clientset.BatchV1().Jobs(namespace)

	_, err = jobs.Create(context.TODO(), jobSpec, metav1.CreateOptions{})

	if err != nil {
		return err
	}

	return nil
}

func (k8s *k8sImpl) AwaitK8sJobDone(jobName string, maxPollCount int, pollIntervalInSeconds int, namespace string) error {
	jobs := k8s.clientset.BatchV1().Jobs(namespace)

	currentPollCount := 0

	for {

		currentPollCount++
		if currentPollCount > maxPollCount {
			duration, err := time.ParseDuration(strconv.Itoa(maxPollCount*pollIntervalInSeconds) + "s")
			if err != nil {
				return err
			}
			return fmt.Errorf("max poll count reached for job %s. Timing out after %s", jobName, duration)
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
				return fmt.Errorf("job %s was suspended. Reason: %s, Message: %s", jobName, condition.Reason, condition.Message)
			case batchv1.JobFailed:
				return fmt.Errorf("job %s failed. Reason: %s, Message: %s", jobName, condition.Reason, condition.Message)
			}
		}

		time.Sleep(time.Duration(pollIntervalInSeconds) * time.Second)
	}
}

// DeleteK8sJob delete a k8s job in the given namespace
func (k8s *k8sImpl) DeleteK8sJob(jobName string, namespace string) error {

	jobs := k8s.clientset.BatchV1().Jobs(namespace)
	return jobs.Delete(context.TODO(), jobName, metav1.DeleteOptions{})
}

func (k8s *k8sImpl) prepareJobEnv(task config.Task, eventData *keptnv2.EventData, jsonEventData interface{}, namespace string) ([]v1.EnvVar, error) {

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

	jobEnv = append(jobEnv,
		v1.EnvVar{
			Name:  "KEPTN_PROJECT",
			Value: eventData.Project,
		},
		v1.EnvVar{
			Name:  "KEPTN_STAGE",
			Value: eventData.Stage,
		},
		v1.EnvVar{
			Name:  "KEPTN_SERVICE",
			Value: eventData.Service,
		},
	)

	return jobEnv, nil
}

func generateEnvFromEvent(env config.Env, jsonEventData interface{}) ([]v1.EnvVar, error) {

	value, err := jsonpath.Get(env.Value, jsonEventData)
	if err != nil {
		return nil, fmt.Errorf("could not add env with name '%v', value '%v', valueFrom '%v': %v", env.Name, env.Value, env.ValueFrom, err)
	}

	if strings.EqualFold(env.Formatting, "yaml") {
		yamlString, err := yaml.Marshal(value)

		if err != nil {
			return nil, fmt.Errorf("could not convert env with name '%v', value '%v', valueFrom '%v' to YAML: %v", env.Name, env.Value, env.ValueFrom, err)
		}

		value = string(yamlString[:])
	} else if strings.EqualFold(env.Formatting, "json") || reflect.ValueOf(value).Kind() == reflect.Map {
		jsonString, err := json.Marshal(value)

		if err != nil {
			return nil, fmt.Errorf("could not convert env with name '%v', value '%v', valueFrom '%v' to JSON: %v", env.Name, env.Value, env.ValueFrom, err)
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

func (k8s *k8sImpl) generateEnvFromSecret(env config.Env, namespace string) ([]v1.EnvVar, error) {

	var generatedEnv []v1.EnvVar

	secret, err := k8s.clientset.CoreV1().Secrets(namespace).Get(context.TODO(), env.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not add env with name %v, valueFrom %v: %v", env.Name, env.ValueFrom, err)
	}

	for key := range secret.Data {
		generatedEnv = append(generatedEnv, v1.EnvVar{
			Name: key,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: env.Name},
					Key:                  key,
				},
			},
		})
	}

	return generatedEnv, nil
}
