package k8sutils

import (
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	keptnutils "github.com/keptn/kubernetes-utils/pkg"
	"k8s.io/client-go/kubernetes"
	"keptn-sandbox/job-executor-service/pkg/config"
)

// k8sImpl is used to interact with kubernetes jobs
type k8sImpl struct {
	clientset kubernetes.Interface
	namespace string
}

//go:generate mockgen -source=connect.go -destination=fake/connect_mock.go -package=fake Interface

// K8s is used to interact with kubernetes jobs
type K8s interface {
	ConnectToCluster() error
	CreateK8sJob(jobName string, action *config.Action, task config.Task, eventData *keptnv2.EventData, configurationServiceURL string, configurationServiceToken string, initContainerImage string, jsonEventData interface{}) error
	DeleteK8sJob(jobName string) error
	GetLogsOfPod(jobName string) (string, error)
}

// NewK8s creates and returns new K8s
func NewK8s(namespace string) K8s {
	return &k8sImpl{namespace: namespace}
}

// ConnectToCluster returns the k8s Clientset
func (k8s *k8sImpl) ConnectToCluster() error {

	config, err := keptnutils.GetClientset(true)
	if err != nil {
		return err
	}
	k8s.clientset = config

	return nil
}
