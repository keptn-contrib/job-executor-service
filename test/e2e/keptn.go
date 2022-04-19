package e2e

import (
	"encoding/base64"
	"fmt"
	"github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils"
	"net/http"
)

const authHeaderName = "x-token"
const protocolScheme = "http"

const jobResourceUri = "job/config.yaml"

type KeptnAPI struct {
	httpClient      *http.Client
	APIHandler      *api.APIHandler
	ProjectHandler  *api.ProjectHandler
	ResourceHandler *api.ResourceHandler
	EventHandler    *api.EventHandler
}

func NewKeptAPI(details KeptnConnectionDetails) KeptnAPI {
	httpClient := http.Client{}

	return KeptnAPI{
		httpClient:      &httpClient,
		APIHandler:      api.NewAuthenticatedAPIHandler(details.Endpoint, details.APIToken, authHeaderName, &httpClient, protocolScheme),
		ProjectHandler:  api.NewAuthenticatedProjectHandler(details.Endpoint, details.APIToken, authHeaderName, &httpClient, protocolScheme),
		ResourceHandler: api.NewAuthenticatedResourceHandler(details.Endpoint, details.APIToken, authHeaderName, &httpClient, protocolScheme),
		EventHandler:    api.NewAuthenticatedEventHandler(details.Endpoint, details.APIToken, authHeaderName, &httpClient, protocolScheme),
	}
}

func (k KeptnAPI) CreateProject(projectName string, shipyardYAML []byte) error {

	shipyardFileBase64 := base64.StdEncoding.EncodeToString(shipyardYAML)

	_, err := k.APIHandler.CreateProject(models.CreateProject{
		Name:     &projectName,
		Shipyard: &shipyardFileBase64,
	})

	if err != nil {
		return fmt.Errorf("unable to create project: %s", logKeptnModelError(err))
	}

	return nil
}

func (k KeptnAPI) DeleteProject(projectName string) error {
	_, err := k.APIHandler.DeleteProject(models.Project{
		ProjectName: projectName,
	})

	if err != nil {
		return fmt.Errorf("unable to delete project: %s", logKeptnModelError(err))
	}

	return nil
}

func (k KeptnAPI) CreateService(projectName string, serviceName string) error {
	_, err := k.APIHandler.CreateService(projectName, models.CreateService{
		ServiceName: &serviceName,
	})

	if err != nil {
		return fmt.Errorf("unable to create service %s in project %s: %s", serviceName, projectName, logKeptnModelError(err))
	}

	return nil
}

func (k KeptnAPI) CreateJobConfig(projectName string, stageName string, serviceName string, jobConfigYaml []byte) error {
	return k.AddServiceResource(projectName, stageName, serviceName, jobResourceUri, string(jobConfigYaml))
}

func (k KeptnAPI) AddServiceResource(projectName string, stageName string, serviceName string, path string, data string) error {
	_, err := k.ResourceHandler.CreateServiceResources(projectName, stageName, serviceName, []*models.Resource{
		{
			Metadata:        nil,
			ResourceContent: data,
			ResourceURI:     &path,
		},
	})

	if err != nil {
		return fmt.Errorf("unable to create service resource for service %s in project %s: %s", serviceName, projectName, err)
	}

	return nil
}

func (k KeptnAPI) SendEvent(keptnEvent *models.KeptnContextExtendedCE) (*models.EventContext, error) {
	keptnContext, err := k.APIHandler.SendEvent(*keptnEvent)

	if err != nil {
		return nil, fmt.Errorf("unable to send event: %s", logKeptnModelError(err))
	}

	return keptnContext, nil
}

func (k KeptnAPI) GetEvents(keptnContext *string) ([]*models.KeptnContextExtendedCE, error) {
	eventFilter := api.EventFilter{
		KeptnContext: *keptnContext,
	}

	events, err := k.EventHandler.GetEvents(&eventFilter)
	if err != nil {
		return nil, fmt.Errorf("unable to get events: %s", logKeptnModelError(err))
	}

	return events, nil
}
