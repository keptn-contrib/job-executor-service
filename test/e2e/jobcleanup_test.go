package e2e

import (
	"github.com/keptn/go-utils/pkg/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestJobCleanupWithSmallTTL(t *testing.T) {
	if !isE2ETestingAllowed() {
		t.Skip("Skipping TestJobCleanupWith0TTLMultipleJobs, not allowed by environment")
	}

	testEnv, err := newTestEnvironment(
		"../events/e2e/jobcleanup.triggered.json",
		"../shipyard/e2e/jobcleanup.deployment.yaml",
		"../data/e2e/jobcleanup.config.yaml",
	)

	require.NoError(t, err)

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
		2*time.Minute,
		1*time.Second,
		keptnContext,
		"sh.keptn.event.deployment.finished",
		func(event *models.KeptnContextExtendedCE) bool {
			responseEventData, err := parseKeptnEventData(event)
			require.NoError(t, err)

			t.Log(responseEventData.Message)

			responseEventData.Message = ""

			assert.Equal(t, expectedEventData, *responseEventData)
			return true
		},
	)
}

func TestJobCleanupWith0TTLMultipleJobs(t *testing.T) {
	t.Skip("Skipping TestJobCleanupWith0TTLMultipleJobs since TTL=0 is not currently supported/working")

	if !isE2ETestingAllowed() {
		t.Skip("Skipping TestJobCleanupWith0TTLMultipleJobs, not allowed by environment")
	}

	testEnv, err := newTestEnvironment(
		"../events/e2e/jobcleanup.triggered.json",
		"../shipyard/e2e/jobcleanup.deployment.yaml",
		"../data/e2e/jobcleanup.0ttl.config.yaml",
	)
	require.NoError(t, err)

	err = testEnv.SetupTestEnvironment()
	require.NoError(t, err)

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

	// If the started event was sent by the job executor we wait for a .finished with the following data:

	// TODO: This is set to fail because JES can never collect logs of jobs with TTL 0s
	expectedEventData := eventData{
		Project: testEnv.EventData.Project,
		Result:  "pass",
		Service: testEnv.EventData.Service,
		Stage:   testEnv.EventData.Stage,
		Status:  "succeeded",
	}

	requireWaitForEvent(t,
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

			assert.Equal(t, expectedEventData, *responseEventData)
			return true
		},
	)
}
