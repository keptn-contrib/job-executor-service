package main

import (
	"context"
	"fmt"
	api "github.com/keptn/go-utils/pkg/api/utils/v2"
	"github.com/keptn/go-utils/pkg/sdk"
	"github.com/sirupsen/logrus"
	"log"
	"os"

	v1 "k8s.io/api/core/v1"

	"keptn-contrib/job-executor-service/pkg/config"
	"keptn-contrib/job-executor-service/pkg/utils"

	"keptn-contrib/job-executor-service/pkg/eventhandler"
	"keptn-contrib/job-executor-service/pkg/k8sutils"

	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
	"github.com/kelseyhightower/envconfig"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	keptn_interface "keptn-contrib/job-executor-service/pkg/keptn"
)

var keptnOptions = keptn.KeptnOpts{}
var env envConfig

type envConfig struct {
	// Port on which to listen for cloudevents
	Port int `envconfig:"RCV_PORT" default:"8080"`
	// Path to which cloudevents are sent
	Path string `envconfig:"RCV_PATH" default:"/"`
	// Whether we are running locally (e.g., for testing) or on production
	Env string `envconfig:"ENV" default:"local"`
	// The k8s namespace the job will run in
	JobNamespace string `envconfig:"JOB_NAMESPACE" required:"true"`
	// URL of the Keptn API endpoint
	KeptnAPIENDPOINT string `envconfig:"KEPTN_API_ENDPOINT" required:"true"`
	// The token of the keptn API
	KeptnAPIToken string `envconfig:"KEPTN_API_TOKEN"`
	// The init container image to use
	InitContainerImage string `envconfig:"INIT_CONTAINER_IMAGE"`
	// Default resource limits cpu for job and init container
	DefaultResourceLimitsCPU string `envconfig:"DEFAULT_RESOURCE_LIMITS_CPU"`
	// Default resource limits memory for job and init container
	DefaultResourceLimitsMemory string `envconfig:"DEFAULT_RESOURCE_LIMITS_MEMORY"`
	// Default resource requests cpu for job and init container
	DefaultResourceRequestsCPU string `envconfig:"DEFAULT_RESOURCE_REQUESTS_CPU"`
	// Default resource requests memory for job and init container
	DefaultResourceRequestsMemory string `envconfig:"DEFAULT_RESOURCE_REQUESTS_MEMORY"`
	// The name of the default job service account which should be used
	DefaultJobServiceAccount string `envconfig:"DEFAULT_JOB_SERVICE_ACCOUNT"`
	// A list of all allowed images that can be used in jobs
	AllowedImageList string `envconfig:"ALLOWED_IMAGE_LIST"  default:""`
	// A flag if privileged job workloads should be allowed by the job-executor-context
	AllowPrivilegedJobs bool `envconfig:"ALLOW_PRIVILEGED_JOBS"`
	// TaskDeadlineSeconds set to an integer > 0 represents the max duration of a task run,
	// a value of 0 allows tasks run for as long as needed (no deadline)
	TaskDeadlineSeconds int64 `envconfig:"TASK_DEADLINE_SECONDS"`
	// FullDeploymentName is the name of the kubernetes deployment of the job executor service,
	// it is used in managed-by labels for jobs and pods that are started by the service
	FullDeploymentName string `envconfig:"FULL_DEPLOYMENT_NAME"`
}

// ServiceName specifies the current services name (e.g., used as source when sending CloudEvents)
const ServiceName = "job-executor-service"

// jobSecurityContextFilePath describes the path of the security config file that is defined in the deployment.yaml
const jobSecurityContextFilePath = "/config/job-defaultSecurityContext.json"

// podSecurityContextFilePath describes the path of the pod security config file that is defined in the deployment.yaml
const podSecurityContextFilePath = "/config/job-podSecurityContext.json"

// jobLabelFilePath describes the path of the job labels yaml file
const jobLabelFilePath = "/config/job-labels.yaml"

// DefaultResourceRequirements contains the default k8s resource requirements for the job and initcontainer, parsed on
// startup from env (treat as const)
var /* const */ DefaultResourceRequirements *v1.ResourceRequirements

// DefaultJobSecurityContext contains the default security context for jobs that are started by the job-executor-service
// the default configuration can be overwritten by job specific configuration
var /* const */ DefaultJobSecurityContext *v1.SecurityContext

// DefaultPodSecurityContext contains the default pod security context for jobs
var /* const */ DefaultPodSecurityContext *v1.PodSecurityContext

// JobLabels contains all user defined labels that should added to each job
var /* const */ JobLabels map[string]string

// TaskDeadlineSecondsPtr represents the max duration of a task run, no limit if nil
var TaskDeadlineSecondsPtr *int64

const serviceName = "job-executor-service"
const eventWildcard = "*"

// NewEventHandler creates a new EventHandler
func NewEventHandler(allowList *utils.ImageFilterList) *eventhandler.EventHandler {
	//uniformHandler := api.NewUniformHandler("localhost:8081/controlPlane")
	return &eventhandler.EventHandler{
		ServiceName: ServiceName,
		Mapper:      new(eventhandler.KeptnCloudEventMapper),
		ImageFilter: imageFilterImpl{
			imageFilterList: allowList,
		},
		JobSettings: k8sutils.JobSettings{
			JobNamespace:                env.JobNamespace,
			KeptnAPIToken:               env.KeptnAPIToken,
			InitContainerImage:          env.InitContainerImage,
			DefaultResourceRequirements: DefaultResourceRequirements,
			DefaultJobServiceAccount:    env.DefaultJobServiceAccount,
			DefaultSecurityContext:      DefaultJobSecurityContext,
			DefaultPodSecurityContext:   DefaultPodSecurityContext,
			AllowPrivilegedJobs:         env.AllowPrivilegedJobs,
			JobLabels:                   JobLabels,
			TaskDeadlineSeconds:         TaskDeadlineSecondsPtr,
			JesDeploymentName:           env.FullDeploymentName,
		},
		K8s: k8sutils.NewK8s(""), // FIXME Why do we pass a namespace if it's ignored?
	}
}

/**
 * Usage: ./main
 * no args: starts listening for cloudnative events on localhost:port/path
 */
func main() {
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
	}

	var err error
	DefaultResourceRequirements, err = k8sutils.CreateResourceRequirements(
		env.DefaultResourceLimitsCPU,
		env.DefaultResourceLimitsMemory,
		env.DefaultResourceRequestsCPU,
		env.DefaultResourceRequestsMemory,
	)
	if err != nil {
		log.Fatalf("unable to create default resource requirements: %v", err.Error())
	}

	if env.AllowPrivilegedJobs {
		log.Println("WARNING: Privileged job workloads are allowed!")
	}

	DefaultJobSecurityContext, err = utils.ReadDefaultJobSecurityContext(jobSecurityContextFilePath)
	if err != nil {
		log.Fatalf("unable to read default job security context: %v", err.Error())
	}

	DefaultPodSecurityContext, err = utils.ReadDefaultPodSecurityContext(podSecurityContextFilePath)
	if err != nil {
		log.Fatalf("unable to read default pod security context: %v", err.Error())
	}

	err = utils.VerifySecurityContext(DefaultPodSecurityContext, DefaultJobSecurityContext, env.AllowPrivilegedJobs)
	if err != nil {
		log.Fatalf("Failed to verify security context: %s", err.Error())
	}

	JobLabels, err = utils.ReadAndValidateJobLabels(jobLabelFilePath)
	if err != nil {
		log.Fatalf("Failed to read job labels: %s", err.Error())
	}

	if env.TaskDeadlineSeconds > 0 {
		TaskDeadlineSecondsPtr = &env.TaskDeadlineSeconds
	}

	_main(os.Args[1:], env)
}

/**
 * Opens up a listener on localhost:port/path and passes incoming requets to gotEvent
 */
func _main(args []string, env envConfig) {
	// Checking if the given job service account is empty
	if env.DefaultJobServiceAccount == "" {
		log.Println("WARNING: No default service account for jobs configured: using kubernetes default service account!")
	}

	keptnOptions.ConfigurationServiceURL = fmt.Sprintf("%s/resource-service", env.KeptnAPIENDPOINT)

	log.Println("Starting job-executor-service...")
	log.Printf("    on Port = %d; Path=%s", env.Port, env.Path)

	ctx := context.Background()
	ctx = cloudevents.WithEncodingStructured(ctx)

	log.Printf("Creating new http handler")

	imageFilterList, err := utils.BuildImageAllowList(env.AllowedImageList)
	if err != nil {
		log.Fatalf("failed to generate the allowlist, %v", err)
	}

	// Handle all events and filter them later in jobEventFilter
	log.Fatal(sdk.NewKeptn(
		serviceName,
		sdk.WithTaskHandler(
			eventWildcard,
			NewEventHandler(imageFilterList),
			jobEventFilter),
		sdk.WithLogger(logrus.New()),
	).Start())
}

// jobEventFilter checks if a job/config exists and whether it contains the event type
func jobEventFilter(keptnHandle sdk.IKeptn, event sdk.KeptnEvent) bool {
	keptnHandle.Logger().Infof("Received event of type: %s from %s with id: %s", *event.Type, *event.Source, event.ID)

	data := &keptnv2.EventData{}
	if err := keptnv2.Decode(event.Data, data); err != nil {
		keptnHandle.Logger().Errorf("Could not parse event: %s", err.Error())
		return false
	}

	jcr := &config.JobConfigReader{
		Keptn: keptn_interface.NewV1ResourceHandler(*data, keptnHandle.APIV2().Resources()),
	}

	// Check if the job configuration can be found
	configuration, _, err := jcr.GetJobConfig(event.GitCommitID)

	if err != nil {
		// TODO: Send error log event
		uniformHandler := api.NewUniformHandler(fmt.Sprintf("%s/controlPlane/v1", env.KeptnAPIENDPOINT))
		errorSender := keptn_interface.NewErrorLogSender(ServiceName, uniformHandler, keptnHandle.APIV2().API())

		errorLogErr := errorSender.SendErrorLogEvent(&event, fmt.Errorf(
			"could not retrieve config for job-executor-service: %w", err,
		))

		if errorLogErr != nil {
			keptnHandle.Logger().Infof("Failed sending error log for keptn context %s: %v. Initial error: %v", event.Shkeptncontext,
				errorLogErr, err)
		}

		keptnHandle.Logger().Infof("could not retrieve config for job-executor-service: %e", err)
		return false
	}

	// Check if we have an action that we can use, if it isn't the case we produce a log output
	mapper := new(eventhandler.KeptnCloudEventMapper)
	eventAsInterface, err := mapper.Map(event)
	if err != nil {
		keptnHandle.Logger().Infof("failed to convert incoming cloudevent: %v", err)
		return false
	}

	hasMatchingEvent := configuration.IsEventMatch(*event.Type, eventAsInterface)
	if !hasMatchingEvent {
		keptnHandle.Logger().Infof(
			"No match found for event %s of type %s. Skipping...", event.ID,
			event.Type,
		)
		return false
	}

	return true
}

type imageFilterImpl struct {
	eventhandler.ImageFilter
	imageFilterList *utils.ImageFilterList
}

func (f imageFilterImpl) IsImageAllowed(image string) bool {
	return f.imageFilterList.Contains(image)
}
