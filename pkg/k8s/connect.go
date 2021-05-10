package k8s

import (
	keptnutils "github.com/keptn/kubernetes-utils/pkg"
	"k8s.io/client-go/kubernetes"
)

func ConnectToCluster(namespace string) (*kubernetes.Clientset, error) {

	config, err := keptnutils.GetClientset(true)
	if err != nil {
		return nil, err
	}

	return config, nil
}
