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

		printInitContainerLogs := false
		containerStatus := getTerminatedContainersWithStatusOfPod(pod)

		// Go through all containers and check if the termination reason is Completed, if not we found a container
		// that exited with an error and therefore have to include all logs from init container, as files could not
		// have been copied over
		for _, container := range containerStatus {
			if container.status.Reason != "Completed" {
				printInitContainerLogs = true
				break
			}
		}

		// Query all logs from containers that have terminated and therefore already had the chance to
		// produce logs, otherwise the k8s api will return an error
		for _, container := range containerStatus {

			// If we don't want to print the init container logs, we just skip this iteration of the
			// loop
			if container.containerType == initContainerType && !printInitContainerLogs {
				continue
			}

			// Query logs of the current selected container
			logsOfContainer, err := getLogsOfContainer(k8s, pod, namespace, container.name)
			if err != nil {
				// In case we can't query the logs of a container, we append the reason instead of the container logs
				logsOfContainer = fmt.Sprintf("Unable to query logs of container: %s", err.Error())
			}

			// Build the final logging output for the container
			logs.WriteString(buildLogOutputForContainer(container, logsOfContainer))
			logs.WriteString("\n")
		}
	}

	return logs.String(), nil
}

const (
	// Indicates that the container is an Init container
	initContainerType = iota
	// Indicates that the container is a container defined in the job workload
	jobContainerType
)

type containerStatus struct {
	name          string
	containerType int
	status        *v1.ContainerStateTerminated
}

// getLogsOfContainer returns the logs of a specific container inside the given pod
func getLogsOfContainer(k8s *K8sImpl, pod v1.Pod, namespace string, container string) (string, error) {

	// Request logs of a specific container
	req := k8s.clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &v1.PodLogOptions{
		Container: container,
	})

	// Stream logs into a buffer
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", err
	}

	defer podLogs.Close()

	// Convert the buffer into a string
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// getTerminatedContainersWithStatusOfPod collects the terminated states of all containers inside a given Pod
func getTerminatedContainersWithStatusOfPod(pod v1.Pod) []containerStatus {
	var containerStatusList []containerStatus

	// Loop over all initContainers in the Pod spec and look at the appropriate
	// InitContainerStatus index to determine the status of the init container
	for index, initContainer := range pod.Spec.InitContainers {
		if pod.Status.InitContainerStatuses[index].State.Terminated != nil {
			containerStatusList = append(containerStatusList, containerStatus{
				name:          initContainer.Name,
				containerType: initContainerType,
				status:        pod.Status.InitContainerStatuses[index].State.Terminated,
			})
		}
	}

	// Loop over all regular containers in the Pod spec and look at the appropriate
	// ContainerStatus index to determine the status of the container
	for index, container := range pod.Spec.Containers {
		if pod.Status.ContainerStatuses[index].State.Terminated != nil {
			containerStatusList = append(containerStatusList, containerStatus{
				name:          container.Name,
				containerType: jobContainerType,
				status:        pod.Status.ContainerStatuses[index].State.Terminated,
			})
		}
	}

	return containerStatusList
}

// buildLogOutputForContainer generates a pretty output of the given logs and the container status in the following
// format. Depending on the status the output changes slightly (output will be empty of no logs are produced):
//
// - Normal output:
// 		Container <container.name>:
//  	<logsOfContainer>
//
// - In case of an error:
// 		Container <container.name> terminated with an error (Reason: <reason> [, Message: <message> |, ExitCode: <code>]):
// 		<logsOfContainer>
//
func buildLogOutputForContainer(container containerStatus, logsOfContainer string) string {
	var logs strings.Builder

	// If the container did not put out any logs, we skip it entirely to prevent polluting the
	// log output too much by appending <no logs available for container> for each container
	if logsOfContainer == "" {
		return ""
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
	logs.WriteString(logsOfContainer)
	logs.WriteString("\n")

	return logs.String()
}
