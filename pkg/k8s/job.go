package k8s

import (
	"context"
	"didiladi/keptn-generic-job-service/pkg/config"
	"fmt"
	"strings"
	"time"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateK8sJob creates a k8s job with the keptn-generic-job-service-initcontainer and the job image of the task and waits until the job finishes
func CreateK8sJob(clientset *kubernetes.Clientset, namespace string, jobName string, action *config.Action, task config.Task, eventData *keptnv2.EventData, configurationServiceURL string, configurationServiceToken string) error {

	var backOffLimit int32 = 0

	jobVolumeName := "job-volume"

	// TODO configure from outside:
	jobVolumeMountPath := "/keptn"

	// TODO configure from outside:
	quantity := resource.MustParse("20Mi")

	// TODO resource quotas from outside

	emptyDirVolume := v1.EmptyDirVolumeSource{
		Medium:    v1.StorageMediumDefault,
		SizeLimit: &quantity,
	}
	automountServiceAccountToken := false

	runAsNonRoot := true
	convert := func(s int64) *int64 {
		return &s
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
						RunAsGroup:   convert(3000),
						FSGroup:      convert(2000),
						RunAsNonRoot: &runAsNonRoot,
					},
					InitContainers: []v1.Container{
						{
							Name:            "init-" + jobName,
							Image:           "yeahservice/keptn-generic-job-service-initcontainer",
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
									Value: configurationServiceURL,
								},
								{
									Name:  "KEPTN_API_TOKEN",
									Value: configurationServiceToken,
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
						},
					},
					Containers: []v1.Container{
						{
							Name:    jobName,
							Image:   task.Image,
							Command: strings.Split(task.Cmd, " "),
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      jobVolumeName,
									MountPath: jobVolumeMountPath,
								},
							},
							Env: []v1.EnvVar{
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
							},
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
				},
			},
			BackoffLimit: &backOffLimit,
		},
	}

	jobs := clientset.BatchV1().Jobs(namespace)

	job, err := jobs.Create(context.TODO(), jobSpec, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// TODO timeout from outside
	// 60 times with 5 seconds wait time => 5 minutes
	const maxPollCount = 60
	const pollIntervalInSeconds = 5
	currentPollCount := 0

	for {

		currentPollCount++
		if currentPollCount > maxPollCount {
			return fmt.Errorf("max poll count reached for job %s. Timing out after 5 minutes", jobName)
		}

		job, err = jobs.Get(context.TODO(), job.Name, metav1.GetOptions{})
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

		time.Sleep(pollIntervalInSeconds * time.Second)
	}

}

// DeleteK8sJob delete a k8s job in the given namespace
func DeleteK8sJob(clientset *kubernetes.Clientset, namespace string, jobName string) error {

	jobs := clientset.BatchV1().Jobs(namespace)
	return jobs.Delete(context.TODO(), jobName, metav1.DeleteOptions{})
}
