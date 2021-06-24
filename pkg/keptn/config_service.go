package keptn

import (
	"fmt"
	"github.com/keptn/go-utils/pkg/api/models"
	"github.com/spf13/afero"
	"net/url"
	"os"
	"strings"
)

//go:generate mockgen -source=config_service.go -destination=fake/config_service_mock.go -package=keptn ConfigService

// ConfigService provides methods to retrieve and match resources from the keptn configuration service
type ConfigService interface {
	GetKeptnResource(fs afero.Fs, resource string) ([]byte, error)
	GetAllKeptnResources(fs afero.Fs, resource string) (map[string][]byte, error)
}

//go:generate mockgen -source=config_service.go -destination=fake/config_service_mock.go -package=keptn ResourceHandler

// ResourceHandler provides methods to work with the keptn configuration service
type ResourceHandler interface {
	GetServiceResource(project string, stage string, service string, resourceURI string) (*models.Resource, error)
	GetAllServiceResources(project string, stage string, service string) ([]*models.Resource, error)
}

type configServiceImpl struct {
	useLocalFileSystem bool
	project            string
	stage              string
	service            string
	resourceHandler    ResourceHandler
}

// NewConfigService creates and returns new ConfigService
func NewConfigService(useLocalFileSystem bool, project string, stage string, service string, resourceHandler ResourceHandler) ConfigService {
	return &configServiceImpl{
		useLocalFileSystem: useLocalFileSystem,
		project:            project,
		stage:              stage,
		service:            service,
		resourceHandler:    resourceHandler,
	}
}

// GetKeptnResource returns a resource from the configuration repo based on the incoming cloud events project, service and stage
func (k *configServiceImpl) GetKeptnResource(fs afero.Fs, resource string) ([]byte, error) {

	// if we run in a runlocal mode we are just getting the file from the local disk
	if k.useLocalFileSystem {
		return k.getKeptnResourceFromLocal(fs, resource)
	}

	// get it from KeptnBase
	// https://github.com/keptn/keptn/issues/2707
	requestedResource, err := k.resourceHandler.GetServiceResource(k.project, k.stage, k.service, url.QueryEscape(resource))

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

	// get it from KeptnBase
	requestedResources, err := k.resourceHandler.GetAllServiceResources(k.project, k.stage, k.service)
	if err != nil {
		return nil, fmt.Errorf("resources not found: %s", err)
	}

	keptnResources := make(map[string][]byte)
	for _, serviceResource := range requestedResources {
		// match against with and without starting slash
		resourceURIWithoutSlash := strings.Replace(*serviceResource.ResourceURI, "/", "", 1)
		if strings.HasPrefix(*serviceResource.ResourceURI, resource) || strings.HasPrefix(resourceURIWithoutSlash, resource) {
			keptnResourceContent, err := k.GetKeptnResource(fs, *serviceResource.ResourceURI)
			if err != nil {
				return nil, fmt.Errorf("could not find file %s", *serviceResource.ResourceURI)
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
	err := afero.Walk(fs, resource, func(path string, info os.FileInfo, err error) error {
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
	})

	if err != nil {
		return nil, err
	}

	return resources, nil
}
