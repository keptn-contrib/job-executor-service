package e2e

import (
	"github.com/keptn/go-utils/pkg/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestHelloWorldDeployment(t *testing.T) {
	if !isE2ETestingAllowed() {
		t.Skip("Skipping TestHelloWorldDeployment, not allowed by environment")
	}

	// Setup the E2E test environment
	testEnv := setupE2ETTestEnvironment(t,
		"../events/e2e.jes.triggered.json",
		"../shipyard/e2e.deployment.yaml",
		"../data/helloworld.config.yaml",
	)

	// Make sure project is delete after the tests are completed
	defer testEnv.CleanupFunc()

	// Send the event to keptn
	keptnContext, err := testEnv.API.SendEvent(testEnv.Event)
	require.NoError(t, err)

	// Checking if the job executor service responded with a .started event
	waitForEvent(t,
		testEnv.API,
		2*time.Minute,
		1*time.Second,
		keptnContext,
		"sh.keptn.event.deployment.started",
		func(_ *models.KeptnContextExtendedCE) bool {
			return true
		},
	)

	t.Log("Received .started event")

	// If the started event was sent by the job executor we wait for a .finished with the following data:
	expectedEventData := eventData{
		Project: testEnv.EventData.Project,
		Result:  "pass",
		Service: testEnv.EventData.Service,
		Stage:   testEnv.EventData.Stage,
		Status:  "succeeded",
	}

	waitForEvent(t,
		testEnv.API,
		5*time.Minute,
		1*time.Second,
		keptnContext,
		"sh.keptn.event.deployment.finished",
		func(event *models.KeptnContextExtendedCE) bool {
			responseEventData, err := parseKeptnEventData(event)
			require.NoError(t, err)

			// If the log contains the Hello world output from the job, we assume that the log
			// was correctly read from the job container and set it as expected message
			if strings.Contains(responseEventData.Message, "Hello World") {
				expectedEventData.Message = responseEventData.Message
			}

			assert.Equal(t, expectedEventData, *responseEventData)
			return true
		},
	)
}
