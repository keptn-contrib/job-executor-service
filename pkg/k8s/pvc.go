package k8s

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateK8sPvc(clientset *kubernetes.Clientset, namespace string, pvcName string, storageClassName string) error {

	pvcs := clientset.CoreV1().PersistentVolumeClaims(namespace)

	filesystem := v1.PersistentVolumeFilesystem
	pvcSpec := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			VolumeMode: &filesystem,
			Resources: v1.ResourceRequirements{
				Requests: map[v1.ResourceName]resource.Quantity{
					v1.ResourceMemory: {
						Format: "100Mi",
					},
				},
			},
			StorageClassName: &storageClassName,
			Selector: metav1.SetAsLabelSelector(map[string]string{
				"keptn-service": "generic-job-service",
			}),
		},
	}

	_, err := pvcs.Create(context.TODO(), pvcSpec, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil

	/*

	// 24 times with 5 seconds wait time -> 2 minutes
	const maxPollCount = 24
	const pollIntervalInSeconds = 5

	currentPollCount := 0

	for {

		currentPollCount++
		if currentPollCount > maxPollCount {
			return fmt.Errorf("max poll count reaced for pvc %s. Timing out after 5 minutes", pvcName)
		}

		pvc, err := pvcs.Get(pvcName, metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				Kind: "pvc",
			},
		})

		if err != nil {
			return err
		}

		pvc.Status

		time.Sleep(pollIntervalInSeconds * time.Second)
	}
	*/
}
