package main

import (
	"didiladi/keptn-generic-job-service/pkg/file"
	"didiladi/keptn-generic-job-service/pkg/keptn"
	"github.com/kelseyhightower/envconfig"
	api "github.com/keptn/go-utils/pkg/api/utils"
	"github.com/spf13/afero"
	"log"
	"os"
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
	resourceHandler := api.NewResourceHandler(env.ConfigurationServiceUrl)
	useLocalFileSystem := false

	// configure keptn options
	if env.Env == "local" {
		log.Println("env=local: Running with local filesystem to fetch resources")
		useLocalFileSystem = true
	}

	configService := keptn.NewKeptnConfigService(useLocalFileSystem, env.Project, env.Stage, env.Service, resourceHandler)

	err := file.MountFiles(env.Action, env.Task, fs, configService)
	if err != nil {
		log.Printf("Error while copying files: %s", err.Error())
		os.Exit(-1)
	}

	os.Exit(0)
}


