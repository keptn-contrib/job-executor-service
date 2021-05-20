package k8s

import (
	keptnutils "github.com/keptn/kubernetes-utils/pkg"
	"k8s.io/client-go/kubernetes"
)

// ConnectToCluster returns the k8s Clientset
func ConnectToCluster() (*kubernetes.Clientset, error) {

	config, err := keptnutils.GetClientset(true)
	if err != nil {
		return nil, err
	}

	return config, nil
}
