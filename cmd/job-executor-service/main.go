package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	api "github.com/keptn/go-utils/pkg/api/utils"
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
	// URL of the Keptn configuration service (this is where we can fetch files from the config repo)
	ConfigurationServiceURL string `envconfig:"CONFIGURATION_SERVICE" default:""`
	// The k8s namespace the job will run in
	JobNamespace string `envconfig:"JOB_NAMESPACE" required:"true"`
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
	// Respond with .finished event if no configuration found
	AlwaysSendFinishedEvent bool `envconfig:"ALWAYS_SEND_FINISHED_EVENT"`
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

/**
 * This method gets called when a new event is received from the Keptn Event Distributor
 * Depending on the Event Type will call the specific event handler functions, e.g: handleDeploymentFinishedEvent
 * See https://github.com/keptn/spec/blob/0.2.0-alpha/cloudevents.md for details on the payload
 */
func processKeptnCloudEvent(ctx context.Context, event cloudevents.Event, allowList *utils.ImageFilterList) error {
	// create keptn handler
	log.Printf("Initializing Keptn Handler")
	myKeptn, err := keptnv2.NewKeptn(&event, keptnOptions)
	if err != nil {
		return fmt.Errorf("could not create Keptn Handler: %w", err)
	}

	log.Printf("gotEvent(%s): %s - %s", event.Type(), myKeptn.KeptnContext, event.Context.GetID())

	// create a uniform handler talking to the distributor
	uniformHandler := api.NewUniformHandler("localhost:8081/controlPlane")
	var eventHandler = &eventhandler.EventHandler{
		Keptn: myKeptn,
		JobConfigReader: &config.JobConfigReader{
			Keptn: keptn_interface.NewV1ResourceHandler(myKeptn.Event, myKeptn.ResourceHandler),
		},
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
			AlwaysSendFinishedEvent:     env.AlwaysSendFinishedEvent,
			DefaultJobServiceAccount:    env.DefaultJobServiceAccount,
			DefaultSecurityContext:      DefaultJobSecurityContext,
			DefaultPodSecurityContext:   DefaultPodSecurityContext,
			AllowPrivilegedJobs:         env.AllowPrivilegedJobs,
			JobLabels:                   JobLabels,
			TaskDeadlineSeconds:         TaskDeadlineSecondsPtr,
			JesDeploymentName:           env.FullDeploymentName,
		},
		K8s:         k8sutils.NewK8s(""), // FIXME Why do we pass a namespoace if it's ignored?
		ErrorSender: keptn_interface.NewErrorLogSender(ServiceName, uniformHandler, myKeptn),
	}

	// prevent duplicate events - https://github.com/keptn/keptn/issues/3888
	go eventHandler.HandleEvent()

	return nil
}

/**
 * Usage: ./main
 * no args: starts listening for cloudnative events on localhost:port/path
 *
 * Environment Variables
 * env=runlocal   -> will fetch resources from local drive instead of configuration service
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

	os.Exit(_main(os.Args[1:], env))
}

/**
 * Opens up a listener on localhost:port/path and passes incoming requets to gotEvent
 */
func _main(args []string, env envConfig) int {

	// configure keptn options
	if env.Env == "local" {
		log.Println("env=local: Running with local filesystem to fetch resources")
		keptnOptions.UseLocalFileSystem = true
	}

	// Checking if the given job service account is empty
	if env.DefaultJobServiceAccount == "" {
		log.Println("WARNING: No default service account for jobs configured: using kubernetes default service account!")
	}

	keptnOptions.ConfigurationServiceURL = env.ConfigurationServiceURL

	log.Println("Starting job-executor-service...")
	log.Printf("    on Port = %d; Path=%s", env.Port, env.Path)

	ctx := context.Background()
	ctx = cloudevents.WithEncodingStructured(ctx)

	log.Printf("Creating new http handler")

	// configure http server to receive cloudevents
	listeningAddr := fmt.Sprintf("localhost:%d", env.Port)
	listener, err := net.Listen(
		"tcp", listeningAddr,
	)
	if err != nil {
		log.Fatalf("error listening on tcp %s: %v", listeningAddr, err)
	}

	p, err := cloudevents.NewHTTP(
		cloudevents.WithPath(env.Path),
		cloudevents.WithListener(
			listener,
		),
	)

	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	imageFilterList, err := utils.BuildImageAllowList(env.AllowedImageList)
	if err != nil {
		log.Fatalf("failed to generate the allowlist, %v", err)
	}

	processCloudEventFunc := func(ctx context.Context, event cloudevents.Event) error {
		return processKeptnCloudEvent(ctx, event, imageFilterList)
	}

	log.Printf("Starting receiver")
	log.Fatal(c.StartReceiver(ctx, processCloudEventFunc))

	return 0
}

type imageFilterImpl struct {
	eventhandler.ImageFilter
	imageFilterList *utils.ImageFilterList
}

func (f imageFilterImpl) IsImageAllowed(image string) bool {
	return f.imageFilterList.Contains(image)
}
