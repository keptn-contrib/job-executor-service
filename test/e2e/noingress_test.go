package e2e

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func TestJobExecutorServiceNoIngress(t *testing.T) {
	if !isE2ETestingAllowed() {
		t.Skipf("Skipping %s, not allowed by environment", t.Name())
	}

	testEnv, err := newTestEnvironment(
		"../events/e2e/noingress.triggered.json",
		"../shipyard/e2e/noingress.deployment.yaml",
		"../data/e2e/empty.config.yaml",
	)

	err = testEnv.SetupTestEnvironment()
	require.NoError(t, err)

	// Make sure project is delete after the tests are completed
	defer testEnv.Cleanup()

	// get jes ip address
	jesPodList, err := getJESPodList(testEnv.K8s, testEnv.Namespace)
	require.NoError(t, err, "Error looking for JES pod(s) in namespace %s", testEnv.Namespace)
	require.NotEmptyf(t, jesPodList.Items, "Unable to find any JES pod in namespace %s", testEnv.Namespace)

	v1PodsEndpoint := testEnv.K8s.CoreV1().Pods(testEnv.Namespace)

	testPodLabels := map[string]string{
		"app.kubernetes.io/name":     "jes-e2e-test",
		"app.kubernetes.io/instance": t.Name(),
	}

	selector := labels.NewSelector()

	for k, v := range testPodLabels {
		labelSelector, _ := labels.NewRequirement(
			k, selection.Equals, []string{v},
		)
		selector = selector.Add(*labelSelector)
	}

	// clean up all test pod when exiting the test
	defer func() {
		v1PodsEndpoint.DeleteCollection(
			context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
				LabelSelector: selector.String(),
			},
		)
	}()

	// run pods to try and do a curl to the ip and port
	for _, jesPod := range jesPodList.Items {
		podIP := jesPod.Status.PodIP

		pod, err := v1PodsEndpoint.Create(
			context.Background(), &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					// in case you are wondering the name is also a reference to Young Frankenstein movie
					GenerateName: "jes-knocker",
					Labels:       testPodLabels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "testcontainer", Image: "curlimages/curl", Args: []string{
								"curl", "-Lv", "--fail-with-body",
								fmt.Sprintf("http://%s:%d/", podIP, 8080),
							},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("32M"),
								},
							},
						},
					},
					RestartPolicy: "Never",
				},
			},
			metav1.CreateOptions{},
		)
		require.NoError(t, err, "unable to create knocker pod")
		waitFor := 30 * time.Second
		tick := 100 * time.Millisecond
		completedPod, err := requireWaitForPodToFinish(t, testEnv.K8s, testEnv.Namespace, *pod, waitFor, tick)
		require.NoError(t, err)

		// exit code 7 from curl signals that the connection to the host has failed
		assert.Equal(t, int32(7), completedPod.Status.ContainerStatuses[0].State.Terminated.ExitCode)
	}

}
