package e2e

import (
	"context"
	"github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils/v2"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
	"time"
)

func TestJobConfig(t *testing.T) {
	if !isE2ETestingAllowed() {
		t.Skipf("Skipping %s, not allowed by environment", t.Name())
	}

	/* const */
	var jobConfigURI = "job/config.yaml"

	testEnv, err := newTestEnvironment(
		"../events/e2e/jobconfig.dev-1.triggered.json",
		"../shipyard/e2e/jobconfig.deployment.yaml",
		"../data/e2e/jobconfig.service-1.config.yaml",
	)

	require.NoError(t, err)

	err = testEnv.SetupTestEnvironment()
	require.NoError(t, err)

	// Make sure project is delete after the tests are completed
	defer testEnv.Cleanup()

	// Upload the stage configuration
	stageConfig, err := os.ReadFile("../data/e2e/jobconfig.stage-dev.config.yaml")
	require.NoError(t, err)

	stageScope := api.NewResourceScope()
	stageScope.Project(testEnv.EventData.Project)
	stageScope.Stage("dev")

	_, err = testEnv.API.ResourceHandler.CreateResource(context.Background(), []*models.Resource{
		{
			ResourceContent: string(stageConfig),
			ResourceURI:     &jobConfigURI,
		},
	}, *stageScope, api.ResourcesCreateResourceOptions{})
	require.NoError(t, err)

	// Upload the project configuration
	projectConfig, err := os.ReadFile("../data/e2e/jobconfig.project.config.yaml")
	require.NoError(t, err)

	projectScope := api.NewResourceScope()
	projectScope.Project(testEnv.EventData.Project)

	_, err = testEnv.API.ResourceHandler.CreateResource(context.Background(), []*models.Resource{
		{
			ResourceContent: string(projectConfig),
			ResourceURI:     &jobConfigURI,
		},
	}, *projectScope, api.ResourcesCreateResourceOptions{})
	require.NoError(t, err)

	tests := []struct {
		Name           string
		Event          string
		ExpectedResult string
	}{
		{
			Name:           "Service job configuration",
			Event:          "../events/e2e/jobconfig.dev-1.triggered.json",
			ExpectedResult: "Hallo Welt",
		},
		{
			Name:           "Stage job configuration",
			Event:          "../events/e2e/jobconfig.dev-2.triggered.json",
			ExpectedResult: "Buon Giorno",
		},
		{
			Name:           "Project job configuration",
			Event:          "../events/e2e/jobconfig.prod.triggered.json",
			ExpectedResult: "Hello World",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			keptnEvent, err := readKeptnContextExtendedCE(test.Event)
			require.NoError(t, err)

			keptnContext, err := testEnv.API.SendEvent(keptnEvent)
			require.NoError(t, err)

			requireWaitForEvent(t,
				testEnv.API,
				1*time.Minute,
				1*time.Second,
				keptnContext,
				"sh.keptn.event.deployment.finished",
				func(event *models.KeptnContextExtendedCE) bool {
					responseEventData, err := parseKeptnEventData(event)
					require.NoError(t, err)

					return strings.Contains(responseEventData.Message, test.ExpectedResult)
				},
			)
		})
	}
}
