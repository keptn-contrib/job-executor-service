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

	testEnv := setupE2ETTestEnvironment(t,
		"../events/e2e.jes.triggered-sleep.json",
		"../shipyard/e2e.deployment.yaml",
		"../data/jobcleanup.config.yaml",
	)

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
	if !isE2ETestingAllowed() {
		t.Skip("Skipping TestJobCleanupWith0TTLMultipleJobs, not allowed by environment")
	}

	testEnv := setupE2ETTestEnvironment(t,
		"../events/e2e.jes.triggered-sleep.json",
		"../shipyard/e2e.deployment.yaml",
		"../data/jobcleanup.0ttl.config.yaml",
	)

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

	// If the started event was sent by the job executor we wait for a .finished with the following data:

	// TODO: This is set to fail because JES can never collect logs of jobs with TTL 0s
	expectedEventData := eventData{
		Project: testEnv.EventData.Project,
		Result:  "pass",
		Service: testEnv.EventData.Service,
		Stage:   testEnv.EventData.Stage,
		Status:  "succeeded",
	}

	waitForEvent(t,
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
