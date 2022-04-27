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

	testEnv, err := newTestEnvironment(
		"../events/e2e/locust.qa.test.triggered.json",
		"../shipyard/e2e/locust.deployment.yaml",
		"../data/e2e/locust.config.yaml",
	)

	require.NoError(t, err)

	err = testEnv.SetupTestEnvironment()
	require.NoError(t, err)

	// Make sure project is delete after the tests are completed
	defer testEnv.Cleanup()

	// Files to upload:
	files := map[string]string{
		"../data/e2e/locust.basic.py": "locust/basic.py",
		"../data/e2e/locust.conf":     "locust/locust.conf",
	}

	for sourceFilePath, resourceURI := range files {
		sourceFileContent, err := ioutil.ReadFile(sourceFilePath)
		require.NoError(t, err)

		err = testEnv.API.AddServiceResource(testEnv.EventData.Project, testEnv.EventData.Stage, testEnv.EventData.Service,
			resourceURI, string(sourceFileContent))

		require.NoError(t, err)
	}

	// Send the event to keptn
	keptnContext, err := testEnv.API.SendEvent(testEnv.Event)
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
