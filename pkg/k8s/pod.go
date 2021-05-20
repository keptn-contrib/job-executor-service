package k8s

import (
	"bytes"
	"context"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetLogsOfPod returns the k8s logs of a job in a namespace
func GetLogsOfPod(clientset *kubernetes.Clientset, namespace string, jobName string) (string, error) {

	// TODO include the logs of the initcontainer

	podLogOpts := v1.PodLogOptions{}

	list, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "job-name=" + jobName,
	})
	if err != nil {
		return "", err
	}

	logs := ""

	for _, pod := range list.Items {

		req := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &podLogOpts)
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
		logs += buf.String()
	}

	return logs, nil
}
