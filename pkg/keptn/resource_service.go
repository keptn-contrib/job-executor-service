package keptn

import (
	"context"
	"fmt"
	"github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils/v2"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"net/url"
	"strings"
)

// EventProperties represents a set of properties of a given cloud event
type EventProperties struct {
	Project     string
	Stage       string
	Service     string
	GitCommitID string
}

// ResourceHandler is an interface that describes the functions for fetching resources from a service, stage or project level
type ResourceHandler interface {
	GetServiceResource(resource string, gitCommitID string) ([]byte, error)
	GetStageResource(resource string, gitCommitID string) ([]byte, error)
	GetProjectResource(resource string, gitCommitID string) ([]byte, error)
	GetAllKeptnResources(resource string) (map[string][]byte, error)
}

// ResourcesInterface is an interface that describes the Keptn resource API
type ResourcesInterface interface {
	CreateResources(ctx context.Context, project string, stage string, service string, resources []*models.Resource, opts api.ResourcesCreateResourcesOptions) (*models.EventContext, *models.Error)
	CreateProjectResources(ctx context.Context, project string, resources []*models.Resource, opts api.ResourcesCreateProjectResourcesOptions) (string, error)
	UpdateProjectResources(ctx context.Context, project string, resources []*models.Resource, opts api.ResourcesUpdateProjectResourcesOptions) (string, error)
	UpdateServiceResources(ctx context.Context, project string, stage string, service string, resources []*models.Resource, opts api.ResourcesUpdateServiceResourcesOptions) (string, error)
	GetAllStageResources(ctx context.Context, project string, stage string, opts api.ResourcesGetAllStageResourcesOptions) ([]*models.Resource, error)
	GetAllServiceResources(ctx context.Context, project string, stage string, service string, opts api.ResourcesGetAllServiceResourcesOptions) ([]*models.Resource, error)
	GetResource(ctx context.Context, scope api.ResourceScope, opts api.ResourcesGetResourceOptions) (*models.Resource, error)
	DeleteResource(ctx context.Context, scope api.ResourceScope, opts api.ResourcesDeleteResourceOptions) error
	UpdateResource(ctx context.Context, resource *models.Resource, scope api.ResourceScope, opts api.ResourcesUpdateResourceOptions) (string, error)
	CreateResource(ctx context.Context, resource []*models.Resource, scope api.ResourceScope, opts api.ResourcesCreateResourceOptions) (string, error)
}

// V1ResourceHandler is a wrapper around the v1 ResourceHandler of the Keptn API to simplify the
// getting of resources of a given event
type V1ResourceHandler struct {
	Event       EventProperties
	ResourceAPI ResourcesInterface
}

// NewV1ResourceHandler creates a new V1ResourceHandler from a given Keptn event and a V1KeptnResourceHandler
func NewV1ResourceHandler(event keptnv2.EventData, resourceAPI ResourcesInterface) ResourceHandler {
	return V1ResourceHandler{
		Event: EventProperties{
			Project: event.GetProject(),
			Stage:   event.GetStage(),
			Service: event.GetService(),
		},
		ResourceAPI: resourceAPI,
	}
}

//go:generate mockgen -source=resource_service.go -destination=fake/keptn_resourcehandler_mock.go -package=fake KeptnResourceHandler

// V1KeptnResourceHandler represents an interface for the api.ResourceHandler struct of the Keptn API
type V1KeptnResourceHandler interface {
	GetResource(scope api.ResourceScope, options ...api.URIOption) (*models.Resource, error)
	//GetAllServiceResources(project string, stage string, service string) ([]*models.Resource, error)
}

// buildResourceHandlerV1Options builds the URIOption list such that it contains a well formatted gitCommitID
func buildResourceHandlerV1Options(gitCommitID string) api.URIOption {
	var queryParam api.URIOption
	if gitCommitID != "" {
		queryParam = api.AppendQuery(url.Values{
			"gitCommitID": []string{gitCommitID},
		})
	} else {
		queryParam = api.AppendQuery(url.Values{})
	}

	return queryParam
}

// GetServiceResource returns the contents of a resource for a given gitCommitID
func (r V1ResourceHandler) GetServiceResource(resource string, gitCommitID string) ([]byte, error) {
	scope := api.NewResourceScope()
	scope.Service(r.Event.Service)
	scope.Project(r.Event.Project)
	scope.Stage(r.Event.Stage)
	scope.Resource(url.QueryEscape(resource))

	options := api.ResourcesGetResourceOptions{URIOptions: []api.URIOption{
		buildResourceHandlerV1Options(gitCommitID),
	}}
	resourceContent, err := r.ResourceAPI.GetResource(context.Background(), *scope, options)
	if err != nil {
		return nil, fmt.Errorf("unable to get resouce from keptn: %w", err)
	}

	return []byte(resourceContent.ResourceContent), nil
}

// GetProjectResource returns the resource that was defined on project level
func (r V1ResourceHandler) GetProjectResource(resource string, gitCommitID string) ([]byte, error) {
	scope := api.NewResourceScope()
	scope.Project(r.Event.Project)
	scope.Resource(url.QueryEscape(resource))

	options := api.ResourcesGetResourceOptions{URIOptions: []api.URIOption{
		buildResourceHandlerV1Options(gitCommitID),
	}}
	resourceContent, err := r.ResourceAPI.GetResource(context.Background(), *scope, options)
	if err != nil {
		return nil, fmt.Errorf("unable to get resouce from keptn: %w", err)
	}

	return []byte(resourceContent.ResourceContent), nil
}

// GetStageResource returns the resource that was defined in the stage
func (r V1ResourceHandler) GetStageResource(resource string, gitCommitID string) ([]byte, error) {
	scope := api.NewResourceScope()
	scope.Project(r.Event.Project)
	scope.Stage(r.Event.Stage)
	scope.Resource(url.QueryEscape(resource))

	options := api.ResourcesGetResourceOptions{URIOptions: []api.URIOption{
		buildResourceHandlerV1Options(gitCommitID),
	}}
	resourceContent, err := r.ResourceAPI.GetResource(context.Background(), *scope, options)
	if err != nil {
		return nil, fmt.Errorf("unable to get resouce from keptn: %w", err)
	}

	return []byte(resourceContent.ResourceContent), nil
}

// GetAllKeptnResources returns a map of keptn resources (key=URI, value=content) from the configuration repo with
// prefix 'resource' (matched with and without leading '/')
func (r V1ResourceHandler) GetAllKeptnResources(resource string) (map[string][]byte, error) {
	keptnResources := make(map[string][]byte)

	// Check for an exact match in the resources - Resource is a file
	keptnResourceContent, err := r.GetServiceResource(resource, "")
	if err == nil {
		keptnResources[resource] = keptnResourceContent
		return keptnResources, nil
	}

	// NOTE:
	// 	Since no exact file has been found, we have to assume that the given resource is a directory.
	// 	Directories don't really exist in the API, so we have to use a HasPrefix match here

	// Get all files from Keptn to enumerate what is in the directory
	requestedResources, err := r.ResourceAPI.GetAllServiceResources(context.Background(), r.Event.Project, r.Event.Stage, r.Event.Service, api.ResourcesGetAllServiceResourcesOptions{})

	if err != nil {
		return nil, fmt.Errorf("unable to list all resources: %w", err)
	}

	// Create a path from the / and append a / to the end to match only files in that directory
	resourceDirectoryName := resource + "/"
	if !strings.HasPrefix(resourceDirectoryName, "/") {
		resourceDirectoryName = "/" + resourceDirectoryName
	}

	for _, serviceResource := range requestedResources {
		if strings.HasPrefix(*serviceResource.ResourceURI, resourceDirectoryName) {

			scope := api.NewResourceScope()
			scope.Project(r.Event.Project)
			scope.Stage(r.Event.Stage)
			scope.Service(r.Event.Service)
			scope.Resource(url.QueryEscape(*serviceResource.ResourceURI))

			// Query resource with the specified git commit id:
			options := api.ResourcesGetResourceOptions{URIOptions: []api.URIOption{
				buildResourceHandlerV1Options(r.Event.GitCommitID),
			}}
			keptnResource, err := r.ResourceAPI.GetResource(context.Background(), *scope, options)
			if err != nil {
				return nil, fmt.Errorf("unable to fetch resource %s: %w", *serviceResource.ResourceURI, err)
			}

			keptnResources[*serviceResource.ResourceURI] = []byte(keptnResource.ResourceContent)
		}
	}

	return keptnResources, nil
}
