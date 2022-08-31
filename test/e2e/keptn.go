package e2e

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils/v2"
	"log"
	"net/http"
	"net/url"
)

const authHeaderName = "x-token"
const protocolScheme = "http"
const jobResourceURI = "job/config.yaml"

// KeptnAPI structure holds different api handlers for the keptn api such that they can be used more easily
type KeptnAPI struct {
	httpClient      *http.Client
	APIHandler      api.APIInterface
	ProjectHandler  api.ProjectsInterface
	ResourceHandler api.ResourcesInterface
	EventHandler    api.EventsInterface
}

// NewKeptnAPI creates a KeptnAPI structure from KeptnConnectionDetails
func NewKeptnAPI(details KeptnConnectionDetails) (*KeptnAPI, error) {
	httpClient := http.Client{}

	endpointURL, err := url.Parse(details.Endpoint)
	if err != nil {
		return nil, err
	}

	endpointScheme := protocolScheme
	if endpointURL.Scheme != "" {
		endpointScheme = endpointURL.Scheme
	}

	apiOptions := []func(*api.APISet){
		api.WithScheme(endpointScheme),
		api.WithAuthToken(details.APIToken, "x-token"),
	}

	// Create the API from the defined options and the URL
	keptnAPI, err := api.New(endpointURL.String(), apiOptions...)
	if err != nil {
		log.Fatalf("unable to create keptn API: %s", err)
	}

	apiHandler := keptnAPI.API()
	projectHandler := keptnAPI.Projects()
	resourceHandler := keptnAPI.Resources()
	eventHandler := keptnAPI.Events()

	return &KeptnAPI{
		httpClient:      &httpClient,
		APIHandler:      apiHandler,
		ProjectHandler:  projectHandler,
		ResourceHandler: resourceHandler,
		EventHandler:    eventHandler,
	}, nil
}

// CreateProject creates a keptn project from the contents of a shipyard yaml file
func (k KeptnAPI) CreateProject(projectName string, shipyardYAML []byte) error {
	ctx := context.Background()

	shipyardFileBase64 := base64.StdEncoding.EncodeToString(shipyardYAML)

	_, err := k.APIHandler.CreateProject(ctx, models.CreateProject{
		Name:     &projectName,
		Shipyard: &shipyardFileBase64,
	}, api.APICreateProjectOptions{})

	if err != nil {
		return fmt.Errorf("unable to create project: %s", convertKeptnModelToErrorString(err))
	}

	return nil
}

// DeleteProject deletes a project by a given name
func (k KeptnAPI) DeleteProject(projectName string) error {
	_, err := k.APIHandler.DeleteProject(context.Background(), models.Project{
		ProjectName: projectName,
	}, api.APIDeleteProjectOptions{})

	if err != nil {
		return fmt.Errorf("unable to delete project: %s", convertKeptnModelToErrorString(err))
	}

	return nil
}

// CreateService creates a service in a given project
func (k KeptnAPI) CreateService(projectName string, serviceName string) error {
	_, err := k.APIHandler.CreateService(context.Background(), projectName, models.CreateService{
		ServiceName: &serviceName,
	}, api.APICreateServiceOptions{})

	if err != nil {
		return fmt.Errorf("unable to create service %s in project %s: %s", serviceName, projectName, convertKeptnModelToErrorString(err))
	}

	return nil
}

// CreateJobConfig uploads the job configuration for the job-executor-service to a specific service and stage
func (k KeptnAPI) CreateJobConfig(projectName string, stageName string, serviceName string, jobConfigYaml []byte) error {
	return k.AddServiceResource(projectName, stageName, serviceName, jobResourceURI, string(jobConfigYaml))
}

// AddServiceResource uploads a resource to a specific service and stage
func (k KeptnAPI) AddServiceResource(projectName string, stageName string, serviceName string, path string, data string) error {
	_, err := k.ResourceHandler.CreateResources(context.Background(), projectName, stageName, serviceName, []*models.Resource{
		{
			Metadata:        nil,
			ResourceContent: data,
			ResourceURI:     &path,
		},
	}, api.ResourcesCreateResourcesOptions{})

	if err != nil {
		return fmt.Errorf("unable to create service resource for service %s in project %s: %s", serviceName, projectName, *err.Message)
	}

	return nil
}

// SendEvent sends an event to Keptn
func (k KeptnAPI) SendEvent(keptnEvent *models.KeptnContextExtendedCE) (*models.EventContext, error) {
	keptnContext, err := k.APIHandler.SendEvent(context.Background(), *keptnEvent, api.APISendEventOptions{})

	if err != nil {
		return nil, fmt.Errorf("unable to send event: %s", convertKeptnModelToErrorString(err))
	}

	return keptnContext, nil
}

// GetEvents returns a list of events for the given context from keptn
func (k KeptnAPI) GetEvents(keptnContext *string) ([]*models.KeptnContextExtendedCE, error) {
	eventFilter := api.EventFilter{
		KeptnContext: *keptnContext,
	}

	events, err := k.EventHandler.GetEvents(context.Background(), &eventFilter, api.EventsGetEventsOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get events: %s", convertKeptnModelToErrorString(err))
	}

	return events, nil
}
