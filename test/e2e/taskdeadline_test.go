package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/keptn/go-utils/pkg/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// TestJobDeadline tests that a task that takes 30 seconds to complete
// (sleeping) succeeds or fails with different settings of taskDeadlineSeconds
func TestJobDeadline(t *testing.T) {
	if !isE2ETestingAllowed() {
		t.Skipf("Skipping %s, not allowed by environment", t.Name())
	}

	testEnv, err := newTestEnvironment(
		"../events/e2e/taskdeadline.triggered.json",
		"../shipyard/e2e/taskdeadline.deployment.yaml",
		"../data/e2e/taskdeadline.config.yaml",
	)

	err = testEnv.SetupTestEnvironment()
	require.NoError(t, err)

	// Make sure project is delete after the tests are completed
	defer testEnv.Cleanup()

	// set a bogus deadline value to retrieve the value before any test is executed
	initialTaskDeadline, err := setConfigMap(
		testEnv.K8s, testEnv.Namespace, "job-service-config", "task_deadline_seconds",
		"",
	)

	require.NoError(t, err)

	defer func() {
		setConfigMap(
			testEnv.K8s, testEnv.Namespace, "job-service-config", "task_deadline_seconds",
			initialTaskDeadline,
		)
		restartJESPod(testEnv.K8s, testEnv.Namespace)
	}()

	tests := []struct {
		name               string
		taskDeadline       string
		expectedCloudEvent eventData
	}{
		{
			name:         "No deadline set - job completes successfully",
			taskDeadline: "0",
			expectedCloudEvent: eventData{
				Project: testEnv.EventData.Project,
				Result:  "pass",
				Service: testEnv.EventData.Service,
				Stage:   testEnv.EventData.Stage,
				Status:  "succeeded",
			},
		},
		{
			name:         "Deadline too short - job fails",
			taskDeadline: "10",
			expectedCloudEvent: eventData{
				Project: testEnv.EventData.Project,
				Result:  "fail",
				Service: testEnv.EventData.Service,
				Stage:   testEnv.EventData.Stage,
				Status:  "errored",
			},
		},
		{
			name:         "Deadline long enough - job succeeds",
			taskDeadline: "60",
			expectedCloudEvent: eventData{
				Project: testEnv.EventData.Project,
				Result:  "pass",
				Service: testEnv.EventData.Service,
				Stage:   testEnv.EventData.Stage,
				Status:  "succeeded",
			},
		},
	}

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				_, err := setConfigMap(
					testEnv.K8s, testEnv.Namespace, "job-service-config", "task_deadline_seconds",
					test.taskDeadline,
				)

				require.NoError(t, err)

				err = restartJESPod(testEnv.K8s, testEnv.Namespace)

				require.NoError(t, err)

				// Send the event to keptn
				keptnContext, err := testEnv.API.SendEvent(testEnv.Event)
				require.NoError(t, err)

				// Checking if the job executor service responded with a .started event
				requireWaitForEvent(
					t,
					testEnv.API,
					2*time.Minute,
					1*time.Second,
					keptnContext,
					"sh.keptn.event.deployment.started",
					func(_ *models.KeptnContextExtendedCE) bool {
						return true
					},
				)

				requireWaitForEvent(
					t,
					testEnv.API,
					2*time.Minute,
					1*time.Second,
					keptnContext,
					"sh.keptn.event.deployment.finished",
					func(event *models.KeptnContextExtendedCE) bool {
						responseEventData, err := parseKeptnEventData(event)
						require.NoError(t, err)

						t.Log(responseEventData.Message)

						responseEventData.Message = ""

						assert.Equal(t, test.expectedCloudEvent, *responseEventData)
						return true
					},
				)
			},
		)
	}

}

func restartJESPod(clientset *kubernetes.Clientset, namespace string) error {

	pods, err := getJESPodList(clientset, namespace)

	if err != nil {
		return fmt.Errorf("unable to list JES pods in namespace %s: %w", namespace, err)
	}

	for _, pod := range pods.Items {
		clientset.CoreV1().Pods(namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
	}
	return nil
}

func setConfigMap(
	clientset *kubernetes.Clientset, namespace string, configMapName string, key string, value string,
) (string,
	error) {

	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(
		context.Background(), configMapName, metav1.GetOptions{},
	)

	if err != nil {
		return "", fmt.Errorf("failed getting configmap %s in namespace %s: %w", configMap, namespace, err)
	}

	prevValue := configMap.Data[key]
	configMap.Data[key] = value

	_, err = clientset.CoreV1().ConfigMaps(namespace).Update(
		context.Background(), configMap, metav1.UpdateOptions{},
	)

	return prevValue, err
}
