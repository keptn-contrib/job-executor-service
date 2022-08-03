package keptn

import (
	"fmt"
	"github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	"net/url"
)

// ResourceHandler is an interface that describes the functions for fetching resources from a service, stage or project level
type ResourceHandler interface {
	// GetServiceResource fetches the specified resource from the state of a given git commit id
	GetServiceResource(resource string, gitCommitID string) ([]byte, error)
	// GetStageResource fetches the specified resource from the state of a given git commit id
	GetStageResource(resource string, gitCommitID string) ([]byte, error)
	// GetProjectResource fetches the specified resource from the project scope
	GetProjectResource(resource string) ([]byte, error)
}

// V1ResourceHandler is a wrapper around the v1 ResourceHandler of the Keptn API to simplify the
// getting of resources of a given event
type V1ResourceHandler struct {
	Event           EventProperties
	ResourceHandler V1KeptnResourceHandler
}

// NewV1ResourceHandler creates a new V1ResourceHandler from a given Keptn event and a V1KeptnResourceHandler
func NewV1ResourceHandler(event keptn.EventProperties, handler V1KeptnResourceHandler) ResourceHandler {
	return V1ResourceHandler{
		Event: EventProperties{
			Project: event.GetProject(),
			Stage:   event.GetStage(),
			Service: event.GetService(),
		},
		ResourceHandler: handler,
	}
}

//go:generate mockgen -source=resource_service.go -destination=fake/keptn_resourcehandler_mock.go -package=fake KeptnResourceHandler

// V1KeptnResourceHandler represents an interface for the api.ResourceHandler struct of the Keptn API
type V1KeptnResourceHandler interface {
	GetResource(scope api.ResourceScope, options ...api.URIOption) (*models.Resource, error)
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
	scope.Resource(resource)

	resourceContent, err := r.ResourceHandler.GetResource(*scope, buildResourceHandlerV1Options(gitCommitID))
	if err != nil {
		return nil, fmt.Errorf("unable to get resouce from keptn: %w", err)
	}

	return []byte(resourceContent.ResourceContent), nil
}

// GetProjectResource returns the resource that was defined on project level
func (r V1ResourceHandler) GetProjectResource(resource string) ([]byte, error) {
	scope := api.NewResourceScope()
	scope.Project(r.Event.Project)
	scope.Resource(resource)

	resourceContent, err := r.ResourceHandler.GetResource(*scope)
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
	scope.Resource(resource)

	resourceContent, err := r.ResourceHandler.GetResource(*scope, buildResourceHandlerV1Options(gitCommitID))
	if err != nil {
		return nil, fmt.Errorf("unable to get resouce from keptn: %w", err)
	}

	return []byte(resourceContent.ResourceContent), nil
}
