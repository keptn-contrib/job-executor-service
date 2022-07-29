package keptn

import (
	"context"
	"fmt"
	"github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils/v2"
	"keptn-contrib/job-executor-service/pkg/config"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/afero"
)

//go:generate mockgen -source=config_service.go -destination=fake/config_service_mock.go -package=fake ConfigService

// ConfigService provides methods to retrieve and match resources from the keptn resource service
type ConfigService interface {
	GetKeptnResource(fs afero.Fs, resource string) ([]byte, error)
	GetAllKeptnResources(fs afero.Fs, resource string) (map[string][]byte, error)

	// GetJobConfiguration returns the fetched job configuration
	GetJobConfiguration(fs afero.Fs) (*config.Config, error)
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
	jobConfigReader    config.JobConfigReader
}

// EventProperties represents a set of properties of a given cloud event
type EventProperties struct {
	Project     string
	Stage       string
	Service     string
	GitCommitID string
}

// NewConfigService creates and returns new ConfigService
func NewConfigService(useLocalFileSystem bool, event EventProperties, resourceHandler V2ResourceHandler) ConfigService {
	configurationService := new(configServiceImpl)

	configurationService.useLocalFileSystem = useLocalFileSystem
	configurationService.eventProperties = event
	configurationService.resourceHandler = resourceHandler
	configurationService.jobConfigReader = config.JobConfigReader{
		Keptn: configurationService,
	}

	return configurationService
}

// buildResourceHandlerV2Options builds the URIOption list such that it contains a well formatted gitCommitID
func buildResourceHandlerV2Options(gitCommitID string) api.ResourcesGetResourceOptions {
	options := api.ResourcesGetResourceOptions{}

	if gitCommitID != "" {
		options.URIOptions = []api.URIOption{
			api.AppendQuery(url.Values{
				"gitCommitID": []string{gitCommitID},
			}),
		}
	}

	return options
}

// fetchKeptnResource sets the resource in the scope correctly and fetches the resource from Keptn with the given gitCommitID
func (k *configServiceImpl) fetchKeptnResource(resource string, scope *api.ResourceScope, gitCommitID string) ([]byte, error) {
	// NOTE: No idea why, but the API requires a double query escape for a path element and does not accept leading /
	//       while emitting absolute paths in the response ...
	scope.Resource(url.QueryEscape(strings.TrimPrefix(resource, "/")))

	requestedResource, err := k.resourceHandler.GetResource(
		context.Background(),
		*scope,
		buildResourceHandlerV2Options(gitCommitID),
	)

	// return Nil in case resource couldn't be retrieved
	if err != nil || requestedResource.ResourceContent == "" {
		return nil, fmt.Errorf("resource not found: %s - %s", resource, err)
	}

	return []byte(requestedResource.ResourceContent), nil
}

// GetKeptnResource returns a resource from the configuration repo based on the incoming cloud events project, service and stage
func (k *configServiceImpl) GetKeptnResource(fs afero.Fs, resource string) ([]byte, error) {

	// if we run in a runlocal mode we are just getting the file from the local disk
	if k.useLocalFileSystem {
		return k.getKeptnResourceFromLocal(fs, resource)
	}

	scope := api.NewResourceScope()
	scope.Project(k.eventProperties.Project)
	scope.Stage(k.eventProperties.Stage)
	scope.Service(k.eventProperties.Service)

	// finally download the resource:
	return k.fetchKeptnResource(resource, scope, k.eventProperties.GitCommitID)
}

// GetAllKeptnResources returns a map of keptn resources (key=URI, value=content) from the configuration repo with
// prefix 'resource' (matched with and without leading '/')
func (k *configServiceImpl) GetAllKeptnResources(fs afero.Fs, resource string) (map[string][]byte, error) {

	// if we run in a runlocal mode we are just getting the file from the local disk
	if k.useLocalFileSystem {
		return k.getKeptnResourcesFromLocal(fs, resource)
	}

	keptnResources := make(map[string][]byte)

	// Check for an exact match in the resources
	keptnResourceContent, err := k.GetKeptnResource(fs, resource)
	if err == nil {
		keptnResources[resource] = keptnResourceContent
		return keptnResources, nil
	}

	// NOTE:
	// 	Since no exact file has been found, we have to assume that the given resource is a directory.
	// 	Directories don't really exist in the API, so we have to use a HasPrefix match here
	scope := api.NewResourceScope()
	scope.Project(k.eventProperties.Project)
	scope.Stage(k.eventProperties.Stage)
	scope.Service(k.eventProperties.Service)

	// Get all files from Keptn to enumerate what is in the directory
	requestedResources, err := k.resourceHandler.GetAllServiceResources(context.Background(),
		k.eventProperties.Project, k.eventProperties.Stage, k.eventProperties.Service,
		api.ResourcesGetAllServiceResourcesOptions{},
	)
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

			// Query resource with the specified git commit id:
			keptnResourceContent, err := k.GetKeptnResource(fs, *serviceResource.ResourceURI)
			if err != nil {
				return nil, fmt.Errorf("unable to fetch resource %s: %w", *serviceResource.ResourceURI, err)
			}

			keptnResources[*serviceResource.ResourceURI] = keptnResourceContent
		}
	}

	return keptnResources, nil
}

func (k *configServiceImpl) GetServiceResource(resource string, gitCommitID string) ([]byte, error) {
	scope := api.NewResourceScope()
	scope.Project(k.eventProperties.Project)
	scope.Stage(k.eventProperties.Stage)
	scope.Service(k.eventProperties.Service)
	return k.fetchKeptnResource(resource, scope, gitCommitID)
}

func (k *configServiceImpl) GetStageResource(resource string, gitCommitID string) ([]byte, error) {
	scope := api.NewResourceScope()
	scope.Project(k.eventProperties.Project)
	scope.Stage(k.eventProperties.Stage)
	return k.fetchKeptnResource(resource, scope, gitCommitID)
}

func (k *configServiceImpl) GetProjectResource(resource string, gitCommitID string) ([]byte, error) {
	scope := api.NewResourceScope()
	scope.Project(k.eventProperties.Project)
	return k.fetchKeptnResource(resource, scope, gitCommitID)
}

func (k *configServiceImpl) GetJobConfiguration(fs afero.Fs) (*config.Config, error) {

	// TODO: How do we fetch different job configs from the local file system?
	// if we run in a runlocal mode we are just getting the file from the local disk
	if k.useLocalFileSystem {
		content, err := k.getKeptnResourceFromLocal(fs, "job/config.yaml")
		if err != nil {
			return nil, fmt.Errorf("unable to read job configuration file: %w", err)
		}

		return config.NewConfig(content)
	}

	config, _, err := k.jobConfigReader.GetJobConfig(k.eventProperties.GitCommitID)
	return config, err
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
