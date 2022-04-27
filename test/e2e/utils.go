package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/keptn/go-utils/pkg/api/models"
	keptnutils "github.com/keptn/kubernetes-utils/pkg"
	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/common/log"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"strconv"
	"testing"
	"time"
)

// KeptnConnectionDetails contains the endpoint and the API token for Keptn
type KeptnConnectionDetails struct {
	Endpoint string
	APIToken string
}

// readKeptnConnectionDetailsFromEnv parses the environment variables and creates a KeptnConnectionDetails
func readKeptnConnectionDetailsFromEnv() KeptnConnectionDetails {
	return KeptnConnectionDetails{
		Endpoint: os.Getenv("KEPTN_ENDPOINT"),
		APIToken: os.Getenv("KEPTN_API_TOKEN"),
	}
}

// isE2ETestingAllowed checks if the E2E tests are allowed to run by parsing environment variables
func isE2ETestingAllowed() bool {
	boolean, err := strconv.ParseBool(os.Getenv("JES_E2E_TEST"))
	if err != nil {
		return false
	}

	return boolean
}

// convertKeptnModelToErrorString transforms the models.Error structure to an error string
func convertKeptnModelToErrorString(keptnError *models.Error) string {
	if keptnError == nil {
		return ""
	}

	if keptnError.Message != nil {
		return fmt.Sprintf("%d, %s", keptnError.Code, *keptnError.Message)
	}

	return fmt.Sprintf("%d <no error message>", keptnError.Code)
}

// readKeptnContextExtendedCE reads a file from a given path and returnes the parsed models.KeptnContextExtendedCE struct
func readKeptnContextExtendedCE(path string) (*models.KeptnContextExtendedCE, error) {
	fileContents, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	var keptnContextExtendedCE models.KeptnContextExtendedCE
	err = json.Unmarshal(fileContents, &keptnContextExtendedCE)

	if err != nil {
		return nil, fmt.Errorf("unable to parse event: %w", err)
	}

	return &keptnContextExtendedCE, nil
}

// eventData structure contains common fields in the data part of a  models.KeptnContextExtendedCE struct that are needed by E2E tests
type eventData struct {
	Message string `mapstruct:"message,omitempty"`
	Project string `mapstruct:"project,omitempty"`
	Result  string `mapstruct:"result,omitempty"`
	Service string `mapstruct:"service,omitempty"`
	Stage   string `mapstruct:"stage,omitempty"`
	Status  string `mapstruct:"status,omitempty"`
}

// parseKeptnEventData parse the Data field of the models.KeptnContextExtendedCE structure into a form, which is more
// convenient to work with
func parseKeptnEventData(ce *models.KeptnContextExtendedCE) (*eventData, error) {
	var eventData eventData
	cfg := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   &eventData,
		TagName:  "json",
	}
	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create mapstructure decoder: %w", err)
	}

	err = decoder.Decode(ce.Data)
	if err != nil {
		return nil, fmt.Errorf("unable to decode event data: %w", err)
	}

	return &eventData, nil
}

// createK8sSecret creates a k8s secret from a json file and uploads it into the give namespace
func createK8sSecret(ctx context.Context, clientset *kubernetes.Clientset, namespace string, jsonFilePath string) (func(ctx2 context.Context), error) {

	// read the file from the given path
	file, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read secrets file: %w", err)
	}

	// unmarshal the contents from the file, since we are using k8s classes it must be a json
	var secret v1.Secret
	err = json.Unmarshal(file, &secret)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal secrets json: %s", err)
	}

	// create the secret in k8s
	_, err = clientset.CoreV1().Secrets(namespace).Create(ctx, &secret, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to create k8s secret: %w", err)
	}

	// return a function which can be used to delete the secret after the tests have finished
	return func(ctx2 context.Context) {
		err := clientset.CoreV1().Secrets(namespace).Delete(ctx2, secret.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Error("Unable to delete secret!")
		}
	}, nil
}

// testEnvironment structure holds different structures and information that are commonly used by the E2E test environment
type testEnvironment struct {
	K8s       *kubernetes.Clientset
	API       KeptnAPI
	EventData *eventData
	Event     *models.KeptnContextExtendedCE
	Namespace string
	shipyard  []byte
	jobConfig []byte
}

// newTestEnvironment creates the basic e2e test environment, by establishing a connection to keptn and parsing the given
// files and extracting the necessary information form it.
func newTestEnvironment(eventJSONFilePath string, shipyardPath string, jobConfigPath string) (*testEnvironment, error) {

	// Read the namespace where the job executor service is
	jesNamespace := os.Getenv("JES_NAMESPACE")
	if jesNamespace == "" {
		return nil, fmt.Errorf("environment variable JES_NAMESPACE must be defined")
	}

	// Just test if we can connect to the cluster
	clientset, err := keptnutils.GetClientset(false)
	if err != nil {
		return nil, fmt.Errorf("unable to get clientset: %w", err)
	}

	// Create a new Keptn api for the use of the E2E test
	keptnAPI := NewKeptnAPI(readKeptnConnectionDetailsFromEnv())

	// Read the event we want to trigger and extract the project, service and stage
	keptnEvent, err := readKeptnContextExtendedCE(eventJSONFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable parse JSON event file: %w", err)
	}

	eventData, err := parseKeptnEventData(keptnEvent)
	if err != nil {
		return nil, fmt.Errorf("unable parse event data of the JSON event: %w", err)
	}

	// Load shipyard file and create the project in Keptn
	shipyardFile, err := ioutil.ReadFile(shipyardPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read the shipyard file: %w", err)
	}

	// Load the job configuration for the E2E test
	jobConfigYaml, err := ioutil.ReadFile(jobConfigPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read the job configuration file: %w", err)
	}

	return &testEnvironment{
		K8s:       clientset,
		API:       keptnAPI,
		EventData: eventData,
		Event:     keptnEvent,
		Namespace: jesNamespace,
		shipyard:  shipyardFile,
		jobConfig: jobConfigYaml,
	}, nil
}

// SetupTestEnvironment Creates the required Project, Service and uploads the Job configuration to Keptn
func (env testEnvironment) SetupTestEnvironment() error {

	err := env.API.CreateProject(env.EventData.Project, env.shipyard)
	if err != nil {
		return fmt.Errorf("unable to create a project in keptn: %w", err)
	}

	// Create a service in Keptn
	err = env.API.CreateService(env.EventData.Project, env.EventData.Service)
	if err != nil {
		return fmt.Errorf("unable to create a service in keptn: %w", err)
	}

	err = env.API.CreateJobConfig(env.EventData.Project, env.EventData.Stage, env.EventData.Service, env.jobConfig)
	if err != nil {
		return err
	}

	return nil
}

// DeleteProject deletes the project that was created in the test environment
func (env testEnvironment) DeleteProject() error {
	return env.API.DeleteProject(env.EventData.Project)
}

// Cleanup deletes all created Keptn resources / services projects
func (env testEnvironment) Cleanup() error {
	if err := env.DeleteProject(); err != nil {
		return err
	}

	return nil
}

// GetKeptnVersion returns the current version of the Keptn server as semver
func (env testEnvironment) GetKeptnVersion() (*semver.Version, error) {
	metadata, errModel := env.API.APIHandler.GetMetadata()
	if errModel != nil {
		return nil, fmt.Errorf("unable to query keptn metadata: %s", convertKeptnModelToErrorString(errModel))
	}

	keptnVersion, err := semver.NewVersion(metadata.Keptnversion)
	if err != nil {
		return nil, fmt.Errorf("unable to convert keptn version to semver: %w", err)
	}

	return keptnVersion, nil
}

// ShouldRun returns an error if the integration test should be skipped based on the given constraint
func (env testEnvironment) ShouldRun(semverConstraint string) error {
	constraint, err := semver.NewConstraint(semverConstraint)
	if err != nil {
		return err
	}

	keptnVersion, err := env.GetKeptnVersion()
	if err != nil {
		return err
	}

	if !constraint.Check(keptnVersion) {
		return fmt.Errorf("skipping test, Keptn version %s does not satisfy expression %s", keptnVersion, constraint)
	}

	return nil
}

// requireWaitForEvent checks if an event occurred in a specific time frame while polling the event bus of keptn, the eventValidator
// should return true if the desired event was found
func requireWaitForEvent(t *testing.T, api KeptnAPI, waitFor time.Duration, tick time.Duration, keptnContext *models.EventContext, eventType string, eventValidator func(c *models.KeptnContextExtendedCE) bool) {
	checkForEventsToMatch := func() bool {
		events, err := api.GetEvents(keptnContext.KeptnContext)
		require.NoError(t, err)

		// for each event we have to check if the type is the correct one and if
		// the source of the event matches the job executor, if that is the case
		// the event can be checked by the eventValidator
		for _, event := range events {
			if *event.Type == eventType && *event.Source == "job-executor-service" {
				if eventValidator(event) {
					return true
				}
			}
		}

		return false
	}

	// We require waiting for a keptn event, this is useful to exit out tests if no .started event occurred.
	// It doesn't make sense in these cases to wait for a .finished or other .triggered events ...
	require.Eventuallyf(t, checkForEventsToMatch, waitFor, tick, "did not receive keptn event: %s", eventType)
}
