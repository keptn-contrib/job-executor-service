package k8sutils

import (
	"bytes"
	"context"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// GetLogsOfPod returns the k8s logs of a job in a namespace
func (k8s *K8sImpl) GetLogsOfPod(jobName string, namespace string) (string, error) {

	list, err := k8s.clientset.CoreV1().Pods(namespace).List(
		context.TODO(), metav1.ListOptions{
			LabelSelector: "job-name=" + jobName,
		},
	)
	if err != nil {
		return "", err
	}

	var logs strings.Builder

	for _, pod := range list.Items {

		// Query all logs from containers that have terminated and therefore already had the chance to
		// produce logs, otherwise the k8s api will return an error
		for _, container := range getTerminatedContainersWithStatusOfPod(pod) {

			// Query logs of the current selected container
			logsOfContainer, err := getLogsOfContainer(k8s, pod, namespace, container.name)
			if err != nil {
				return "", err
			}

			// Prepend the container name at the beginning, so we are able to separate logs of different containers
			// and display a termination error at the beginning, may be more interesting than the logs of the container
			if container.status.Reason != "Completed" {
				logs.WriteString("Container ")
				logs.WriteString(container.name)
				logs.WriteString(" terminated with an error (Reason: ")
				logs.WriteString(container.status.Reason)

				// Sometimes the message is not given, to provide prettier logs we just don't print the
				// message part if it doesn't exist
				if container.status.Message != "" {
					logs.WriteString(", Message: ")
					logs.WriteString(container.status.Message)
					logs.WriteString(")")
				} else {
					logs.WriteString(", ExitCode: ")
					logs.WriteString(fmt.Sprintf("%d", container.status.ExitCode))
					logs.WriteString(")")
				}

				logs.WriteString(":\n")
			} else {
				logs.WriteString("Container ")
				logs.WriteString(container.name)
				logs.WriteString(":\n")
			}

			// Finally, append the actual logs of the container or a default message to the log
			if logsOfContainer != "" {
				logs.WriteString(logsOfContainer)
				logs.WriteString("\n")
			} else {
				logs.WriteString("<no logs available for container>\n")
			}

			logs.WriteString("\n")
		}
	}

	return logs.String(), nil
}

type containerStatus struct {
	name   string
	status *v1.ContainerStateTerminated
}

func getLogsOfContainer(k8s *K8sImpl, pod v1.Pod, namespace string, container string) (string, error) {
	req := k8s.clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &v1.PodLogOptions{
		Container: container,
	})

	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", err
	}

	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func getTerminatedContainersWithStatusOfPod(pod v1.Pod) []containerStatus {
	var containerStatusList []containerStatus

	for index, initContainer := range pod.Spec.InitContainers {
		if pod.Status.InitContainerStatuses[index].State.Terminated != nil {
			containerStatusList = append(containerStatusList, containerStatus{
				name:   initContainer.Name,
				status: pod.Status.InitContainerStatuses[index].State.Terminated,
			})
		}
	}

	for index, container := range pod.Spec.Containers {
		if pod.Status.ContainerStatuses[index].State.Terminated != nil {
			containerStatusList = append(containerStatusList, containerStatus{
				name:   container.Name,
				status: pod.Status.ContainerStatuses[index].State.Terminated,
			})
		}
	}

	return containerStatusList
}
