package main

import (
	"didiladi/keptn-generic-job-service/pkg/config"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	api "github.com/keptn/go-utils/pkg/api/utils"
	"github.com/spf13/afero"
	"log"
	"os"
	"path/filepath"
)

type envConfig struct {
	// Port on which to listen for cloudevents
	Port int `envconfig:"RCV_PORT" default:"8080"`
	// Path to which cloudevents are sent
	Path string `envconfig:"RCV_PATH" default:"/"`
	// Whether we are running locally (e.g., for testing) or on production
	Env string `envconfig:"ENV" default:"local"`
	// URL of the Keptn configuration service (this is where we can fetch files from the config repo)
	ConfigurationServiceUrl string `envconfig:"CONFIGURATION_SERVICE" default:""`
	// The keptn project contained in the initial cloud event
	Project string `envconfig:"KEPTN_PROJECT" required:"true"`
	// The keptn stage contained in the initial cloud event
	Stage string `envconfig:"KEPTN_STAGE" required:"true"`
	// The keptn service contained in the initial cloud event
	Service string `envconfig:"KEPTN_SERVICE" required:"true"`
	// The name of the config action which triggered the init container run
	Action string `envconfig:"JOB_ACTION" required:"true"`
	// The name of the config task which triggered the init container run
	Task string `envconfig:"JOB_TASK" required:"true"`
}

func main() {

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
	}

	fs := afero.NewOsFs()
	err := mountFiles(env, fs)
	if err != nil {
		log.Printf("Error while copying files: %s", err.Error())
		os.Exit(-1)
	}
	os.Exit(0)
}

func mountFiles(env envConfig, fs afero.Fs) error {

	resourceHandler := api.NewResourceHandler(env.ConfigurationServiceUrl)
	useLocalFileSystem := false

	// configure keptn options
	if env.Env == "local" {
		log.Println("env=local: Running with local filesystem to fetch resources")
		useLocalFileSystem = true
	}

	resource, err := getKeptnResource(env, useLocalFileSystem, resourceHandler, "generic-job/config.yaml")
	if err != nil {
		log.Printf("Could not find config for generic Job service")
		return err
	}

	configuration, err := config.NewConfig(resource)
	if err != nil {
		log.Printf("Could not parse config: %s", err)
		return err
	}

	found, action := configuration.FindActionByName(env.Action)
	if !found {
		return fmt.Errorf("no action found with name '%s'", env.Action)
	}

	found, task := action.FindTaskByName(env.Task)
	if !found {
		return fmt.Errorf("no task found with name '%s'", env.Task)
	}

	for _, fileName := range task.Files {

		resource, err = getKeptnResource(env, useLocalFileSystem, resourceHandler, fileName)
		if err != nil {
			log.Printf("Could not find file %s for task %s", fileName, env.Task)
			return err
		}

		// Our mount starts with /keptn
		dir := filepath.Join("/keptn", filepath.Dir(fileName))
		fullFilePath := filepath.Join("/keptn", fileName)

		err := fs.MkdirAll(dir, 0700)
		if err != nil {
			log.Printf("Could not create directory %s for file %s", dir, fileName)
			return err
		}

		file, err := fs.Create(fullFilePath)
		if err != nil {
			log.Printf("Could not create file %s", fileName)
			return err
		}

		_, err = file.Write(resource)
		defer func() {
			err = file.Close()
			if err != nil {
				log.Printf("Could not close file %s", file.Name())
			}
		}()

		if err != nil {
			log.Printf("Could not write to file %s", fileName)
			return err
		}
	}

	return nil
}

// getKeptnResource returns a resource from the configuration repo based on the incoming cloud events project, service and stage
func getKeptnResource(env envConfig, useLocalFileSystem bool, resourceHandler *api.ResourceHandler, resource string) ([]byte, error) {

	// if we run in a runlocal mode we are just getting the file from the local disk
	if useLocalFileSystem {
		return getKeptnResourceFromLocal(resource)
	}

	// get it from KeptnBase
	requestedResource, err := resourceHandler.GetServiceResource(env.Project, env.Stage, env.Service, resource)

	// return Nil in case resource couldn't be retrieved
	if err != nil || requestedResource.ResourceContent == "" {
		return nil, fmt.Errorf("resource not found: %s - %s", resource, err)
	}

	return []byte(requestedResource.ResourceContent), nil
}

/**
 * Retrieves a resource (=file) from the local file system. Basically checks if the file is available and if so returns it
 */
func getKeptnResourceFromLocal(resource string) ([]byte, error) {
	if _, err := os.Stat(resource); err == nil {
		return []byte(resource), nil
	} else {
		return nil, err
	}
}


