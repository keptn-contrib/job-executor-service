package e2e

import (
	"github.com/keptn/go-utils/pkg/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func TestGitCommitID(t *testing.T) {
	if !isE2ETestingAllowed() {
		t.Skip("Skipping TestHelloWorldDeployment, not allowed by environment")
	}

	// Setup the E2E test environment
	testEnv, err := newTestEnvironment(
		"../events/e2e/helloworld.triggered.json",
		"../shipyard/e2e/helloworld.deployment.yaml",
		"../data/e2e/helloworld.config.yaml",
	)

	require.NoError(t, err)

	err = testEnv.SetupTestEnvironment()
	require.NoError(t, err)

	// Make sure project is delete after the tests are completed
	defer testEnv.Cleanup()

	// Make sure the integration test is only run for Keptn versions that support the
	// gitCommitId parameter for resource queries
	if err := testEnv.ShouldRun(">=0.16.0"); err != nil {
		t.Skipf("%s\n", err.Error())
	}

	// Send the event to keptn
	keptnContext, err := testEnv.API.SendEvent(testEnv.Event)
	require.NoError(t, err)

	// Wait for the deployment to be completed
	var gitCommitId string
	requireWaitForEvent(t,
		testEnv.API,
		5*time.Minute,
		1*time.Second,
		keptnContext,
		"sh.keptn.event.deployment.finished",
		func(event *models.KeptnContextExtendedCE) bool {
			gitCommitId = event.GitCommitID
			return true
		},
	)

	// Upload new job config, that should not get executed when the old git commit id is used:
	jobConfigYaml, err := ioutil.ReadFile("../data/e2e/gitcommitid.config.yaml")
	require.NoError(t, err)

	err = testEnv.API.CreateJobConfig(testEnv.EventData.Project, testEnv.EventData.Stage, testEnv.EventData.Service, jobConfigYaml)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	testEnv.Event.GitCommitID = gitCommitId
	keptnContext, err = testEnv.API.SendEvent(testEnv.Event)
	require.NoError(t, err)

	// Assert that still the old Hello World example is run and not the new one ...
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
