package keptn

import (
	"fmt"
	"github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	"net/url"
)

// V1ResourceHandler is a wrapper around the v1 ResourceHandler of the Keptn API to simplify the
// getting of resources of a given event
type V1ResourceHandler struct {
	Event           EventProperties
	ResourceHandler KeptnResourceHandler
}

func NewV1ResourceHandler(event keptn.EventProperties, handler KeptnResourceHandler) V1ResourceHandler {
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

// KeptnResourceHandler represents an interface for the api.ResourceHandler struct of the Keptn API
type KeptnResourceHandler interface {
	GetResource(scope api.ResourceScope, options ...api.URIOption) (*models.Resource, error)
}

// GetResource returns the contents of a resource for a given gitCommitId
func (r V1ResourceHandler) GetResource(resource string, gitCommitId string) ([]byte, error) {
	scope := api.NewResourceScope()
	scope.Resource(url.QueryEscape(resource))
	scope.Service(r.Event.Service)
	scope.Project(r.Event.Project)
	scope.Stage(r.Event.Stage)

	var queryParam api.URIOption
	if gitCommitId != "" {
		queryParam = api.AppendQuery(url.Values{
			"gitCommitID": []string{gitCommitId},
		})
	} else {
		queryParam = api.AppendQuery(url.Values{})
	}

	resourceContent, err := r.ResourceHandler.GetResource(*scope, queryParam)
	if err != nil {
		return nil, fmt.Errorf("unable to get resouce from keptn: %w", err)
	}

	return []byte(resourceContent.ResourceContent), nil
}
