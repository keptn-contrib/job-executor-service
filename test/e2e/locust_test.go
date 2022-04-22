package e2e

import (
	"github.com/keptn/go-utils/pkg/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
	"time"
)

func TestLocust(t *testing.T) {
	if !isE2ETestingAllowed() {
		t.Skip("Skipping TestLocust, not allowed by environment")
	}

	testEnv := setupE2ETTestEnvironment(t,
		"../events/e2e.locust.qa.test.triggered.json",
		"../shipyard/e2e.locust.yaml",
		"../data/locust.config.yaml",
	)

	defer testEnv.CleanupFunc()

	// Files to upload:
	files := map[string]string{
		"../data/locust.basic.py": "locust/basic.py",
		"../data/locust.conf":     "locust/locust.conf",
	}

	for sourceFilePath, resourceUri := range files {
		sourceFileContent, err := ioutil.ReadFile(sourceFilePath)
		require.NoError(t, err)

		err = testEnv.API.AddServiceResource(testEnv.EventData.Project, testEnv.EventData.Stage, testEnv.EventData.Service,
			resourceUri, string(sourceFileContent))

		require.NoError(t, err)
	}

	// Send the event to keptn
	keptnContext, err := testEnv.API.SendEvent(testEnv.Event)
	require.NoError(t, err)

	// Checking if the job executor service responded with a .started event
	waitForEvent(t,
		testEnv.API,
		2*time.Minute,
		1*time.Second,
		keptnContext,
		"sh.keptn.event.test.started",
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
		"sh.keptn.event.test.finished",
		func(event *models.KeptnContextExtendedCE) bool {
			responseEventData, err := parseKeptnEventData(event)
			require.NoError(t, err)

			// We are not interested in comparing the output of locust
			responseEventData.Message = ""

			assert.Equal(t, expectedEventData, *responseEventData)
			return true
		},
	)
}
