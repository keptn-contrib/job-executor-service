package keptn

import (
	"fmt"
	api "github.com/keptn/go-utils/pkg/api/utils"
	"os"
)

//go:generate mockgen -source=config_service.go -destination=config_service_mock.go -package=keptn KeptnConfigService

type KeptnConfigService interface {
	GetKeptnResource(resource string) ([]byte, error)
}

type keptnConfigServiceImpl struct {
	useLocalFileSystem bool
	project string
	stage string
	service string
	resourceHandler *api.ResourceHandler
}

func NewKeptnConfigService(useLocalFileSystem bool, project string, stage string, service string, resourceHandler *api.ResourceHandler) KeptnConfigService {
	return &keptnConfigServiceImpl{
		useLocalFileSystem: useLocalFileSystem,
		project:            project,
		stage:              stage,
		service:            service,
		resourceHandler:    resourceHandler,
	}
}

// getKeptnResource returns a resource from the configuration repo based on the incoming cloud events project, service and stage
func (k *keptnConfigServiceImpl) GetKeptnResource(resource string) ([]byte, error) {

	// if we run in a runlocal mode we are just getting the file from the local disk
	if k.useLocalFileSystem {
		return k.getKeptnResourceFromLocal(resource)
	}

	// get it from KeptnBase
	requestedResource, err := k.resourceHandler.GetServiceResource(k.project, k.stage, k.service, resource)

	// return Nil in case resource couldn't be retrieved
	if err != nil || requestedResource.ResourceContent == "" {
		return nil, fmt.Errorf("resource not found: %s - %s", resource, err)
	}

	return []byte(requestedResource.ResourceContent), nil
}

/**
 * Retrieves a resource (=file) from the local file system. Basically checks if the file is available and if so returns it
 */
func (k *keptnConfigServiceImpl) getKeptnResourceFromLocal(resource string) ([]byte, error) {
	if _, err := os.Stat(resource); err == nil {
		return []byte(resource), nil
	} else {
		return nil, err
	}
}