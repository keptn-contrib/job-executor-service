package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateK8sJob(clientset *kubernetes.Clientset, namespace string, jobName string, image string, cmd string) error {

	var backOffLimit int32 = 0

	jobVolumeName := "job-volume"
	jobVolumeMountPath := "/mnt"

	quantity := resource.MustParse("20Mi")
	emptyDirVolume := v1.EmptyDirVolumeSource{
		Medium:    v1.StorageMediumDefault,
		SizeLimit: &quantity,
	}
	automountServiceAccountToken := false

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name:    "init-" + jobName,
							Image:   "bash",
							Command: []string{"sh", "-c", "touch /mnt/hello"},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      jobVolumeName,
									MountPath: jobVolumeMountPath,
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:    jobName,
							Image:   "bash",
							Command: strings.Split(cmd, " "),
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      jobVolumeName,
									MountPath: jobVolumeMountPath,
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
			return fmt.Errorf("max poll count reaced for job %s. Timing out after 5 minutes", jobName)
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

func DeleteK8sJob(clientset *kubernetes.Clientset, namespace string, jobName string) error {

	jobs := clientset.BatchV1().Jobs(namespace)
	return jobs.Delete(context.TODO(), jobName, metav1.DeleteOptions{})
}
