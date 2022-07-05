package keptn

import (
	"context"
	"fmt"
	"github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils/v2"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/afero"
)

//go:generate mockgen -source=config_service.go -destination=fake/config_service_mock.go -package=fake ConfigService

// ConfigService provides methods to retrieve and match resources from the keptn configuration service
type ConfigService interface {
	GetKeptnResource(fs afero.Fs, resource string) ([]byte, error)
	GetAllKeptnResources(fs afero.Fs, resource string) (map[string][]byte, error)
}

//go:generate mockgen -source=config_service.go -destination=fake/config_service_mock.go -package=fake V2ResourceHandler

// V2ResourceHandler provides methods to work with the keptn configuration service
type V2ResourceHandler interface {
	// GetAllServiceResources returns a list of all resources.
	GetAllServiceResources(ctx context.Context, project string, stage string, service string,
		opts api.ResourcesGetAllServiceResourcesOptions) ([]*models.Resource, error)

	// GetResource returns a resource from the defined ResourceScope.
	GetResource(ctx context.Context, scope api.ResourceScope, opts api.ResourcesGetResourceOptions) (*models.Resource, error)
}

type configServiceImpl struct {
	useLocalFileSystem bool
	eventProperties    EventProperties
	resourceHandler    V2ResourceHandler
}

type EventProperties struct {
	Project     string
	Stage       string
	Service     string
	GitCommitId string
}

// NewConfigService creates and returns new ConfigService
func NewConfigService(useLocalFileSystem bool, event EventProperties, resourceHandler V2ResourceHandler) ConfigService {
	return &configServiceImpl{
		useLocalFileSystem: useLocalFileSystem,
		eventProperties:    event,
		resourceHandler:    resourceHandler,
	}
}

// GetKeptnResource returns a resource from the configuration repo based on the incoming cloud events project, service and stage
func (k *configServiceImpl) GetKeptnResource(fs afero.Fs, resource string) ([]byte, error) {

	// if we run in a runlocal mode we are just getting the file from the local disk
	if k.useLocalFileSystem {
		return k.getKeptnResourceFromLocal(fs, resource)
	}

	// https://github.com/keptn/keptn/issues/2707
	// Note: trimming the prefix is necessary if a gitCommitId is used (?)
	encodedResource := url.QueryEscape(strings.TrimPrefix(resource, "/"))

	scope := api.NewResourceScope()
	scope.Project(k.eventProperties.Project)
	scope.Stage(k.eventProperties.Stage)
	scope.Service(k.eventProperties.Service)
	scope.Resource(encodedResource)

	options := api.ResourcesGetResourceOptions{}
	if k.eventProperties.GitCommitId != "" {
		options.URIOptions = []api.URIOption{
			api.AppendQuery(url.Values{
				"gitCommitID": []string{k.eventProperties.GitCommitId},
			}),
		}
	}

	requestedResource, err := k.resourceHandler.GetResource(context.Background(), *scope, options)

	log.Printf("%#v\n%#v\n%#v\n", scope, requestedResource, err)

	// return Nil in case resource couldn't be retrieved
	if err != nil || requestedResource.ResourceContent == "" {
		return nil, fmt.Errorf("resource not found: %s - %s", resource, err)
	}

	return []byte(requestedResource.ResourceContent), nil
}

// GetAllKeptnResources returns a map of keptn resources (key=URI, value=content) from the configuration repo with
// prefix 'resource' (matched with and without leading '/')
func (k *configServiceImpl) GetAllKeptnResources(fs afero.Fs, resource string) (map[string][]byte, error) {

	// if we run in a runlocal mode we are just getting the file from the local disk
	if k.useLocalFileSystem {
		return k.getKeptnResourcesFromLocal(fs, resource)
	}

	scope := api.NewResourceScope()
	scope.Project(k.eventProperties.Project)
	scope.Stage(k.eventProperties.Stage)
	scope.Service(k.eventProperties.Service)

	// Get all resources from Keptn in the current service
	requestedResources, err := k.resourceHandler.GetAllServiceResources(context.Background(),
		k.eventProperties.Project, k.eventProperties.Stage, k.eventProperties.Service,
		api.ResourcesGetAllServiceResourcesOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("resources not found: %s", err)
	}

	// Go over the resources and fetch the content of each resource with the given gitCommitId:
	keptnResources := make(map[string][]byte)
	for _, serviceResource := range requestedResources {
		// match against with and without starting slash
		// Note: this makes it possible to include directories, maybe a glob might be a better idea
		resourceURIWithoutSlash := strings.Replace(*serviceResource.ResourceURI, "/", "", 1)
		if strings.HasPrefix(*serviceResource.ResourceURI, resource) || strings.HasPrefix(
			resourceURIWithoutSlash, resource,
		) {
			keptnResourceContent, err := k.GetKeptnResource(fs, *serviceResource.ResourceURI)
			if err != nil {
				return nil, fmt.Errorf("could not find file %s for version %s",
					*serviceResource.ResourceURI, k.eventProperties.GitCommitId,
				)
			}
			keptnResources[*serviceResource.ResourceURI] = keptnResourceContent
		}
	}

	return keptnResources, nil
}

/**
 * Retrieves a resource (=file) from the local file system. Basically checks if the file is available and if so returns it
 */
func (k *configServiceImpl) getKeptnResourceFromLocal(fs afero.Fs, resource string) ([]byte, error) {
	_, err := fs.Stat(resource)
	if err != nil {
		return nil, err
	}

	content, err := afero.ReadFile(fs, resource)
	if err != nil {
		return nil, err
	}
	return content, nil
}

/**
* Retrieves a resource (=file or all files of a directory) from the local file system. Basically checks if the file/directory
  is available and if so returns it or its files
*/
func (k *configServiceImpl) getKeptnResourcesFromLocal(fs afero.Fs, resource string) (map[string][]byte, error) {
	resources := make(map[string][]byte)
	err := afero.Walk(
		fs, resource, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			content, err := k.getKeptnResourceFromLocal(fs, path)
			if err != nil {
				return err
			}
			resources[path] = content
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return resources, nil
}
