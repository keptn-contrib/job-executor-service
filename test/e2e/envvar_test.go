package e2e

import (
	"context"
	"github.com/keptn/go-utils/pkg/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestEnvironmentVariables(t *testing.T) {
	if !isE2ETestingAllowed() {
		t.Skip("Skipping TestEnvironmentVariables, not allowed by environment")
	}

	testEnv := setupE2ETTestEnvironment(t,
		"../events/e2e.jes.triggered-labels.json",
		"../shipyard/e2e.deployment.yaml",
		"../data/envvar.config.yaml",
	)

	defer testEnv.CleanupFunc()

	// Create a secret for the jobs
	secretContext, cancelFunc := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelFunc()

	secretDeleteFunc, err := createK8sSecret(secretContext, testEnv.K8s, testEnv.Namespace, "../data/envvar.secret.json")
	require.NoError(t, err)
	defer secretDeleteFunc(context.Background())

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

	logEntriesRegex, err := regexp.Compile(`([A-Za-z0-9_-]+\s?=.*)\n`)
	require.NoError(t, err)

	expectedLogMessages := []string{
		"LABELS_E2E_LABEL=EXAMPLE_LABEL",
		"LABELS_JSONPATHLABEL=JSON_PATH_LABEL_VALUE",
		"LABELS_E2E_LABEL=EXAMPLE_LABEL",
		"E2E-jsonPathLabel=JSON_PATH_LABEL_VALUE",
		"E2E-EVENT={\"jsonPathArray\":[\"\\u003cnot-displayed\\u003e\",\"\\u003cdisplayed\\u003e\"]}",
		"E2E-jsonPathArray=<displayed>",
		"E2E-Secret=secret-kubernetes-variable",
		"LABELS_E2E_LABEL=EXAMPLE_LABEL",
		"E2E-HOST=https://keptn.sh",
		"LABELS_E2E_LABEL=EXAMPLE_LABEL",
		"E2E_DATA_DIR=/tmp/data",
	}

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

			// get the log messages and check if they are as expected
			logMessages := logEntriesRegex.FindAllString(responseEventData.Message, -1)
			for i := range logMessages {
				logMessages[i] = strings.TrimSpace(logMessages[i])
			}
			assert.Equal(t, expectedLogMessages, logMessages)

			// check if the rest of the event is expected
			responseEventData.Message = ""
			assert.Equal(t, expectedEventData, *responseEventData)

			return true
		},
	)
}
