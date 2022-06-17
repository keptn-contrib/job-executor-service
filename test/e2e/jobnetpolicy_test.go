package e2e

import (
	"context"
	"fmt"
	"github.com/keptn/go-utils/pkg/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestJobNetworkPolicy_NoIngress(t *testing.T) {
	if !isE2ETestingAllowed() {
		t.Skipf("Skipping %s, not allowed by environment", t.Name())
	}

	testEnv, err := newTestEnvironment(
		"../events/e2e/nginx.triggered.json",
		"../shipyard/e2e/nginx.deployment.yaml",
		"../data/e2e/nginx.config.yaml",
	)

	err = testEnv.SetupTestEnvironment()
	require.NoError(t, err)

	// Make sure project is delete after the tests are completed
	defer testEnv.Cleanup()

	// Send the event to keptn
	keptnContext, err := testEnv.API.SendEvent(testEnv.Event)
	require.NoError(t, err)

	// Checking if the job executor service responded with a .started event
	requireWaitForEvent(t,
		testEnv.API,
		2*time.Minute,
		1*time.Second,
		keptnContext,
		"sh.keptn.event.deployment.started",
		func(_ *models.KeptnContextExtendedCE) bool {
			return true
		},
	)

	// Wait for the nginx pod to get ready
	waitFor := 60 * time.Second
	tick := 500 * time.Millisecond
	require.Eventually(t,
		func() bool {
			pods, err := getJESRunningJobPods(testEnv.K8s, testEnv.Namespace)
			assert.NoError(t, err)

			return len(pods) >= 1
		},
		waitFor,
		tick,
		"timed out while waiting job pod to get ready", waitFor,
	)

	jobPods, err := getJESRunningJobPods(testEnv.K8s, testEnv.Namespace)
	require.NoError(t, err)

	for _, targetPod := range jobPods {
		if targetPod.Spec.Containers[0].Command[0] != "nginx" {
			continue
		}

		curlPod, err := testEnv.K8s.CoreV1().Pods(testEnv.Namespace).Create(context.Background(), &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("test-no-ingress-for-%s", targetPod.Name)[:63],
			},
			Spec: v1.PodSpec{
				RestartPolicy: v1.RestartPolicyNever,
				Containers: []v1.Container{
					{
						Name:    "curl-container",
						Image:   "alpine/curl",
						Command: []string{"curl"},
						Args: []string{
							"-Lv",
							"--fail-with-body",
							fmt.Sprintf("http://%s:%d/", targetPod.Status.PodIP, 8080),
						},
					},
				},
			},
		}, metav1.CreateOptions{})
		require.NoError(t, err)

		// Wait for spawned pod to finish
		waitFor := 30 * time.Second
		tick := 100 * time.Millisecond
		completedPod, err := requireWaitForPodToFinish(t, testEnv.K8s, testEnv.Namespace, *curlPod, waitFor, tick)
		require.NoError(t, err)

		expectedCurlExitStatus := int32(0)

		// Check if the network policy is applied or not
		networkPolicies, err := testEnv.K8s.NetworkingV1().NetworkPolicies(testEnv.Namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		for _, netPolicy := range networkPolicies.Items {
			if netPolicy.Name == "jes-job-network-policy" {
				// exit code 7 from curl signals that the connection to the host has failed
				expectedCurlExitStatus = int32(7)
			}
		}
		assert.Equal(t, expectedCurlExitStatus, completedPod.Status.ContainerStatuses[0].State.Terminated.ExitCode)

		// Cleanup the pods
		defer func(targetPod v1.Pod) {
			err = testEnv.K8s.CoreV1().Pods(testEnv.Namespace).Delete(context.Background(), curlPod.Name, metav1.DeleteOptions{})
			assert.NoError(t, err, "unable to delete curl pod")

			err = testEnv.K8s.CoreV1().Pods(testEnv.Namespace).Delete(context.Background(), targetPod.Name, metav1.DeleteOptions{})
			assert.NoError(t, err, "unable to delete nginx pod")
		}(targetPod)
	}
}

func TestJobNetworkPolicy_NoEgress(t *testing.T) {
	if !isE2ETestingAllowed() {
		t.Skipf("Skipping %s, not allowed by environment", t.Name())
	}

	testEnv, err := newTestEnvironment(
		"../events/e2e/job-egress.triggered.json",
		"../shipyard/e2e/job-egress.deployment.yaml",
		"../data/e2e/job-egress.config.yaml",
	)

	err = testEnv.SetupTestEnvironment()
	require.NoError(t, err)

	// Make sure project is delete after the tests are completed
	defer testEnv.Cleanup()

	// In this integration tests we send 3 different events to JES:
	var /*const*/ eventTypeEgressExternTriggered = "sh.keptn.event.e2e.egress-extern.triggered"
	var /*const*/ eventTypeEgressAPIServerTriggered = "sh.keptn.event.e2e.egress-apiserver.triggered"
	var /*const*/ eventTypeEgressK8sTriggered = "sh.keptn.event.e2e.egress-k8s.triggered"

	// Send the event to keptn
	testEnv.Event.Type = &eventTypeEgressExternTriggered
	keptnContext, err := testEnv.API.SendEvent(testEnv.Event)
	require.NoError(t, err)

	// Checking if the job executor service responded with a .started event
	requireWaitForEvent(t,
		testEnv.API,
		2*time.Minute,
		1*time.Second,
		keptnContext,
		"sh.keptn.event.deployment.started",
		func(_ *models.KeptnContextExtendedCE) bool {
			return true
		},
	)

	// If the started event was sent by the job executor we wait for a .finished with the following data:
	expectedEventData := eventData{
		Project: testEnv.EventData.Project,
		Result:  "pass",
		Service: testEnv.EventData.Service,
		Stage:   testEnv.EventData.Stage,
		Status:  "succeeded",
	}

	requireWaitForEvent(t,
		testEnv.API,
		5*time.Minute,
		1*time.Second,
		keptnContext,
		"sh.keptn.event.deployment.finished",
		func(event *models.KeptnContextExtendedCE) bool {
			responseEventData, err := parseKeptnEventData(event)
			require.NoError(t, err)

			responseEventData.Message = ""
			assert.Equal(t, expectedEventData, *responseEventData)
			return true
		},
	)

	// Check if the network policy is applied or not
	networkPolicies, err := testEnv.K8s.NetworkingV1().NetworkPolicies(testEnv.Namespace).List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err)

	var networkPolicy *netv1.NetworkPolicy
	for _, policy := range networkPolicies.Items {
		if policy.Name == "jes-job-network-policy" {
			networkPolicy = &policy
			break
		}
	}

	// Check if a network policy is defined and the access to the keptn api server has been enabled
	if networkPolicy != nil {
		allowingAccessToKeptn := false
		for _, egress := range networkPolicy.Spec.Egress {
			if egress.To[0].PodSelector != nil && egress.To[0].PodSelector.MatchLabels["app.kubernetes.io/name"] == "api-gateway-nginx" {
				allowingAccessToKeptn = true
				break
			}
		}

		if !allowingAccessToKeptn {
			expectedEventData = eventData{
				Project: testEnv.EventData.Project,
				Result:  "fail",
				Service: testEnv.EventData.Service,
				Stage:   testEnv.EventData.Stage,
				Status:  "errored",
			}
		}
	}

	testEnv.Event.Type = &eventTypeEgressAPIServerTriggered
	keptnContext, err = testEnv.API.SendEvent(testEnv.Event)
	require.NoError(t, err)

	// Checking if the job executor service responded with a .started event
	requireWaitForEvent(t,
		testEnv.API,
		2*time.Minute,
		1*time.Second,
		keptnContext,
		"sh.keptn.event.test.started",
		func(_ *models.KeptnContextExtendedCE) bool {
			return true
		},
	)

	requireWaitForEvent(t,
		testEnv.API,
		5*time.Minute,
		1*time.Second,
		keptnContext,
		"sh.keptn.event.test.finished",
		func(event *models.KeptnContextExtendedCE) bool {
			responseEventData, err := parseKeptnEventData(event)
			require.NoError(t, err)

			responseEventData.Message = ""
			assert.Equal(t, expectedEventData, *responseEventData)
			return true
		},
	)

	// Test if the job can access the kubernetes api endpoint, if a network policy is defined this shouldn't be allowed
	if networkPolicy != nil {
		expectedEventData = eventData{
			Project: testEnv.EventData.Project,
			Result:  "fail",
			Service: testEnv.EventData.Service,
			Stage:   testEnv.EventData.Stage,
			Status:  "errored",
		}
	}

	testEnv.Event.Type = &eventTypeEgressK8sTriggered
	keptnContext, err = testEnv.API.SendEvent(testEnv.Event)
	require.NoError(t, err)

	requireWaitForEvent(t,
		testEnv.API,
		2*time.Minute,
		1*time.Second,
		keptnContext,
		"sh.keptn.event.release.started",
		func(_ *models.KeptnContextExtendedCE) bool {
			return true
		},
	)

	requireWaitForEvent(t,
		testEnv.API,
		5*time.Minute,
		1*time.Second,
		keptnContext,
		"sh.keptn.event.release.finished",
		func(event *models.KeptnContextExtendedCE) bool {
			responseEventData, err := parseKeptnEventData(event)
			require.NoError(t, err)

			responseEventData.Message = ""
			assert.Equal(t, expectedEventData, *responseEventData)
			return true
		},
	)

}
