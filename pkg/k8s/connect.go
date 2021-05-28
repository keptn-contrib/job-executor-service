package k8s

import (
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	keptnutils "github.com/keptn/kubernetes-utils/pkg"
	"k8s.io/client-go/kubernetes"
	"keptn-sandbox/job-executor-service/pkg/config"
)

// K8s is used to interact with kubernetes jobs
type K8s struct {
}

//go:generate mockgen -source=connect.go -destination=fake/connect_mock.go -package=fake Interface

// Interface is used to interact with kubernetes jobs
type Interface interface {
	ConnectToCluster() (*kubernetes.Clientset, error)
	CreateK8sJob(clientset *kubernetes.Clientset, namespace string, jobName string, action *config.Action, task config.Task, eventData *keptnv2.EventData, configurationServiceURL string, configurationServiceToken string, initContainerImage string, jsonEventData interface{}) error
	DeleteK8sJob(clientset *kubernetes.Clientset, namespace string, jobName string) error
	GetLogsOfPod(clientset *kubernetes.Clientset, namespace string, jobName string) (string, error)
}

// ConnectToCluster returns the k8s Clientset
func (*K8s) ConnectToCluster() (*kubernetes.Clientset, error) {

	config, err := keptnutils.GetClientset(true)
	if err != nil {
		return nil, err
	}

	return config, nil
}
