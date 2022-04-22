package k8sutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestGetLogsOfPodHappyPath(t *testing.T) {
	k8sClientSet := k8sfake.NewSimpleClientset()

	jobName := "completed-job"
	namespace := "namespace"

	initContainerName := "init-" + jobName
	initContainerImage := "keptncontrib/job-executor-service-initcontainer"
	volumeSize := resource.MustParse("20Mi")

	mainContainerImage := "some-work-image"

	podName := "some-job-pod"
	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:   podName,
			Labels: map[string]string{"job-name": jobName},
		},
		Spec: v1.PodSpec{
			InitContainers: []v1.Container{
				{
					Name:  initContainerName,
					Image: initContainerImage,
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "job-volume",
							ReadOnly:  false,
							MountPath: "/keptn",
						},
					},
					ImagePullPolicy: v1.PullIfNotPresent,
				},
			},
			Containers: []v1.Container{
				{
					Name:    jobName,
					Image:   mainContainerImage,
					Command: []string{"do-work"},
					Args:    []string{"now"},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "job-volume",
							ReadOnly:  false,
							MountPath: "/keptn",
						},
					},
					ImagePullPolicy: v1.PullIfNotPresent,
				},
			},
			ServiceAccountName: "default-job-account",
			Volumes: []v1.Volume{
				{
					Name: "job-volume",
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{
							Medium:    v1.StorageMediumDefault,
							SizeLimit: &volumeSize,
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
			InitContainerStatuses: []v1.ContainerStatus{
				{
					Name: initContainerName,
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode: 0,
							Reason:   "Container terminated normally",
							Message:  "Init done.",
						},
					},
					Image: initContainerImage,
				},
			},
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name: jobName,
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode: 0,
							Reason:   "Container terminated normally",
							Message:  "Work done.",
						},
					},
					Image: mainContainerImage,
				},
			},
		},
	}
	k8sClientSet.CoreV1().Pods(namespace).Create(context.Background(), &pod, metav1.CreateOptions{})

	// this block would be the right signature intercepting a k8s fake client call but doesn't work for logs,
	// sadly this is sidestepped in the k8s client fake impl. See kubernetes/typed/core/v1/fake/fake_pod_expansion.go#GetLogs
	//
	// k8sClientSet.AddReactor(
	// 	"get", "v1/pod", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
	// 		if action.GetSubresource() != "log" {
	// 			return false, nil, nil
	// 		}
	// 		return true, nil, nil
	// 	},
	// )

	k8s := K8sImpl{clientset: k8sClientSet}

	logsOfPod, err := k8s.GetLogsOfPod(jobName, namespace)
	assert.NoError(t, err)
	assert.Contains(t, logsOfPod, "fake logs")

	// Assert that the fake received the call
	getLogActionInitContainer := k8stesting.GenericActionImpl{
		ActionImpl: k8stesting.ActionImpl{
			Namespace: namespace,
			Verb:      "get",
			Resource: schema.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "pods",
			},
			Subresource: "log",
		},
		Value: &v1.PodLogOptions{
			Container: initContainerName,
		},
	}

	getLogActionContainer := k8stesting.GenericActionImpl{
		ActionImpl: k8stesting.ActionImpl{
			Namespace: namespace,
			Verb:      "get",
			Resource: schema.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "pods",
			},
			Subresource: "log",
		},
		Value: &v1.PodLogOptions{
			Container: jobName,
		},
	}

	assert.Contains(t, k8sClientSet.Actions(), getLogActionInitContainer)
	assert.Contains(t, k8sClientSet.Actions(), getLogActionContainer)
}
