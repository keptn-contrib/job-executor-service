package k8sutils

import (
	"context"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"keptn-sandbox/job-executor-service/pkg/github/model"
)

func (k8s *k8sImpl) CreateImageBuilder(jobName string, step model.Step, registry string) (string, error) {

	var backOffLimit int32 = 0

	convert := func(s int64) *int64 {
		return &s
	}

	imageRegistryPath := registry + "/" + step.Uses

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: k8s.namespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					SecurityContext: &v1.PodSecurityContext{
						RunAsUser:  convert(1000),
						RunAsGroup: convert(2000),
						FSGroup:    convert(2000),
					},
					Containers: []v1.Container{
						{
							Name:  jobName,
							Image: "gcr.io/kaniko-project/executor:latest",
							Env: []v1.EnvVar{
								{
									Name:  "GOOGLE_APPLICATION_CREDENTIALS",
									Value: "/kaniko/config.json",
								},
							},
							Args: []string{
								"--destination " + imageRegistryPath,
								"--context git://github.com/" + step.Uses,
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "gcr-secret",
									MountPath: "/kaniko",
								},
								{
									Name:      "workspace",
									MountPath: "/workspace",
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
					Volumes: []v1.Volume{
						{
							Name: "workspace",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "gcr-secret",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "kaniko",
									Items: []v1.KeyToPath{
										{
											Key:  "config.json",
											Path: "config.json",
											Mode: nil,
										},
									},
								},
							},
						},
					},
				},
			},
			BackoffLimit: &backOffLimit,
		},
	}

	jobs := k8s.clientset.BatchV1().Jobs(k8s.namespace)

	_, err := jobs.Create(context.TODO(), jobSpec, metav1.CreateOptions{})

	if err != nil {
		return "", err
	}

	return imageRegistryPath, nil
}
