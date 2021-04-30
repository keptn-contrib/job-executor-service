package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
	"github.com/kelseyhightower/envconfig"
	keptnlib "github.com/keptn/go-utils/pkg/lib"
	keptn "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

var keptnOptions = keptn.KeptnOpts{}

type envConfig struct {
	// Port on which to listen for cloudevents
	Port int `envconfig:"RCV_PORT" default:"8080"`
	// Path to which cloudevents are sent
	Path string `envconfig:"RCV_PATH" default:"/"`
	// Whether we are running locally (e.g., for testing) or on production
	Env string `envconfig:"ENV" default:"local"`
	// URL of the Keptn configuration service (this is where we can fetch files from the config repo)
	ConfigurationServiceUrl string `envconfig:"CONFIGURATION_SERVICE" default:""`
}

// ServiceName specifies the current services name (e.g., used as source when sending CloudEvents)
const ServiceName = "keptn-service-template-go"

/**
 * Parses a Keptn Cloud Event payload (data attribute)
 */
func parseKeptnCloudEventPayload(event cloudevents.Event, data interface{}) error {
	err := event.DataAs(data)
	if err != nil {
		log.Fatalf("Got Data Error: %s", err.Error())
		return err
	}
	return nil
}

/**
 * This method gets called when a new event is received from the Keptn Event Distributor
 * Depending on the Event Type will call the specific event handler functions, e.g: handleDeploymentFinishedEvent
 * See https://github.com/keptn/spec/blob/0.2.0-alpha/cloudevents.md for details on the payload
 */
func processKeptnCloudEvent(ctx context.Context, event cloudevents.Event) error {
	// create keptn handler
	log.Printf("Initializing Keptn Handler")
	myKeptn, err := keptnv2.NewKeptn(&event, keptnOptions)
	if err != nil {
		return errors.New("Could not create Keptn Handler: " + err.Error())
	}

	log.Printf("gotEvent(%s): %s - %s", event.Type(), myKeptn.KeptnContext, event.Context.GetID())

	if err != nil {
		log.Printf("failed to parse incoming cloudevent: %v", err)
		return err
	}

	/**
		* CloudEvents types in Keptn 0.8.0 follow the following pattern:
		* - sh.keptn.event.${EVENTNAME}.triggered
		* - sh.keptn.event.${EVENTNAME}.started
		* - sh.keptn.event.${EVENTNAME}.status.changed
		* - sh.keptn.event.${EVENTNAME}.finished
		*
		* For convenience, types can be generated using the following methods:
		* - triggered:      keptnv2.GetTriggeredEventType(${EVENTNAME}) (e.g,. keptnv2.GetTriggeredEventType(keptnv2.DeploymentTaskName))
		* - started:        keptnv2.GetStartedEventType(${EVENTNAME}) (e.g,. keptnv2.GetStartedEventType(keptnv2.DeploymentTaskName))
		* - status.changed: keptnv2.GetStatusChangedEventType(${EVENTNAME}) (e.g,. keptnv2.GetStatusChangedEventType(keptnv2.DeploymentTaskName))
		* - finished:       keptnv2.GetFinishedEventType(${EVENTNAME}) (e.g,. keptnv2.GetFinishedEventType(keptnv2.DeploymentTaskName))
		*
		* The following Cloud Events are reserved and specified in the Keptn spec:
		* - approval
		* - deployment
		* - test
		* - evaluation
		* - release
		* - remediation
		* - action
		* - get-sli (for quality-gate SLI providers)
		* - problem / problem.open (both deprecated, use action or remediation instead)

		* There are more "internal" Cloud Events that might not have all four status, e.g.:
	    * - project
		* - project.create
		* - service
		* - service.create
		* - configure-monitoring
		*
		* For those Cloud Events the keptn/go-utils library conveniently provides several data structures
		* and strings in github.com/keptn/go-utils/pkg/lib/v0_2_0, e.g.:
		* - deployment: DeploymentTaskName, DeploymentTriggeredEventData, DeploymentStartedEventData, DeploymentFinishedEventData
		* - test: TestTaskName, TestTriggeredEventData, TestStartedEventData, TestFinishedEventData
		* - ... (they all follow the same pattern)
		*
		*
		* In most cases you will be interested in processing .triggered events (e.g., sh.keptn.event.deployment.triggered),
		* which you an achieve as follows:
		* if event.type() == keptnv2.GetTriggeredEventType(keptnv2.DeploymentTaskName) { ... }
		*
		* Processing the event payload can be achieved as follows:
		*
		* eventData := &keptnv2.DeploymentTriggeredEventData{}
		* parseKeptnCloudEventPayload(event, eventData)
		*
		* See https://github.com/keptn/spec/blob/0.2.0-alpha/cloudevents.md for more details of Keptn Cloud Events and their payload
		* Also, see https://github.com/keptn-sandbox/echo-service/blob/a90207bc119c0aca18368985c7bb80dea47309e9/pkg/events.go as an example how to create your own CloudEvents
		**/

	/**
	* The following code presents a very generic implementation of processing almost all possible
	* Cloud Events that are retrieved by this service.
	* Please follow the documentation provided above for more guidance on the different types.
	* Feel free to delete parts that you don't need.
	**/
	switch event.Type() {

	// -------------------------------------------------------
	// sh.keptn.event.project.create - Note: This is due to change
	case keptnv2.GetStartedEventType(keptnv2.ProjectCreateTaskName): // sh.keptn.event.project.create.started
		log.Printf("Processing Project.Create.Started Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.ProjectCreateStartedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)
	case keptnv2.GetFinishedEventType(keptnv2.ProjectCreateTaskName): // sh.keptn.event.project.create.finished
		log.Printf("Processing Project.Create.Finished Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.ProjectCreateFinishedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)
	// -------------------------------------------------------
	// sh.keptn.event.service.create - Note: This is due to change
	case keptnv2.GetStartedEventType(keptnv2.ServiceCreateTaskName): // sh.keptn.event.service.create.started
		log.Printf("Processing Service.Create.Started Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.ServiceCreateStartedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)
	case keptnv2.GetFinishedEventType(keptnv2.ServiceCreateTaskName): // sh.keptn.event.service.create.finished
		log.Printf("Processing Service.Create.Finished Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.ServiceCreateFinishedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)

	// -------------------------------------------------------
	// sh.keptn.event.approval
	case keptnv2.GetTriggeredEventType(keptnv2.ApprovalTaskName): // sh.keptn.event.approval.triggered
		log.Printf("Processing Approval.Triggered Event")

		eventData := &keptnv2.ApprovalTriggeredEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		return HandleApprovalTriggeredEvent(myKeptn, event, eventData)
	case keptnv2.GetStartedEventType(keptnv2.ApprovalTaskName): // sh.keptn.event.approval.started
		log.Printf("Processing Approval.Started Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.ApprovalStartedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)
	case keptnv2.GetFinishedEventType(keptnv2.ApprovalTaskName): // sh.keptn.event.approval.finished
		log.Printf("Processing Approval.Finished Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.ApprovalFinishedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)

	// -------------------------------------------------------
	// sh.keptn.event.deployment
	case keptnv2.GetTriggeredEventType(keptnv2.DeploymentTaskName): // sh.keptn.event.deployment.triggered
		log.Printf("Processing Deployment.Triggered Event")

		eventData := &keptnv2.DeploymentTriggeredEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		return HandleDeploymentTriggeredEvent(myKeptn, event, eventData)
	case keptnv2.GetStartedEventType(keptnv2.DeploymentTaskName): // sh.keptn.event.deployment.started
		log.Printf("Processing Deployment.Started Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.DeploymentStartedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)
	case keptnv2.GetFinishedEventType(keptnv2.DeploymentTaskName): // sh.keptn.event.deployment.finished
		log.Printf("Processing Deployment.Finished Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.DeploymentFinishedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)

	// -------------------------------------------------------
	// sh.keptn.event.test
	case keptnv2.GetTriggeredEventType(keptnv2.TestTaskName): // sh.keptn.event.test.triggered
		log.Printf("Processing Test.Triggered Event")

		eventData := &keptnv2.TestTriggeredEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		return HandleTestTriggeredEvent(myKeptn, event, eventData)
	case keptnv2.GetStartedEventType(keptnv2.TestTaskName): // sh.keptn.event.test.started
		log.Printf("Processing Test.Started Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.TestStartedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)
	case keptnv2.GetFinishedEventType(keptnv2.TestTaskName): // sh.keptn.event.test.finished
		log.Printf("Processing Test.Finished Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.TestFinishedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)

	// -------------------------------------------------------
	// sh.keptn.event.evaluation
	case keptnv2.GetTriggeredEventType(keptnv2.EvaluationTaskName): // sh.keptn.event.evaluation.triggered
		log.Printf("Processing Evaluation.Triggered Event")

		eventData := &keptnv2.EvaluationTriggeredEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		return HandleEvaluationTriggeredEvent(myKeptn, event, eventData)
	case keptnv2.GetStartedEventType(keptnv2.EvaluationTaskName): // sh.keptn.event.evaluation.started
		log.Printf("Processing Evaluation.Started Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.EvaluationStartedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)
	case keptnv2.GetFinishedEventType(keptnv2.EvaluationTaskName): // sh.keptn.event.evaluation.finished
		log.Printf("Processing Evaluation.Finished Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.EvaluationFinishedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)

	// -------------------------------------------------------
	// sh.keptn.event.release
	case keptnv2.GetTriggeredEventType(keptnv2.ReleaseTaskName): // sh.keptn.event.release.triggered
		log.Printf("Processing Release.Triggered Event")

		eventData := &keptnv2.ReleaseTriggeredEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		return HandleReleaseTriggeredEvent(myKeptn, event, eventData)
	case keptnv2.GetStartedEventType(keptnv2.ReleaseTaskName): // sh.keptn.event.release.started
		log.Printf("Processing Release.Started Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.ReleaseStartedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)
	case keptnv2.GetStatusChangedEventType(keptnv2.ReleaseTaskName): // sh.keptn.event.release.status.changed
		log.Printf("Processing Release.Status.Changed Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.ReleaseStatusChangedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)
	case keptnv2.GetFinishedEventType(keptnv2.ReleaseTaskName): // sh.keptn.event.release.finished
		log.Printf("Processing Release.Finished Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.ReleaseFinishedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)

	// -------------------------------------------------------
	// sh.keptn.event.remediation
	case keptnv2.GetTriggeredEventType(keptnv2.RemediationTaskName): // sh.keptn.event.remediation.triggered
		log.Printf("Processing Remediation.Triggered Event")

		eventData := &keptnv2.RemediationTriggeredEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		return HandleRemediationTriggeredEvent(myKeptn, event, eventData)
	case keptnv2.GetStartedEventType(keptnv2.RemediationTaskName): // sh.keptn.event.remediation.started
		log.Printf("Processing Remediation.Started Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.RemediationStartedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)
	case keptnv2.GetStatusChangedEventType(keptnv2.RemediationTaskName): // sh.keptn.event.remediation.status.changed
		log.Printf("Processing Remediation.Status.Changed Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.RemediationStatusChangedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)
	case keptnv2.GetFinishedEventType(keptnv2.RemediationTaskName): // sh.keptn.event.remediation.finished
		log.Printf("Processing Remediation.Finished Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.RemediationFinishedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)

	// -------------------------------------------------------
	// sh.keptn.event.action
	case keptnv2.GetTriggeredEventType(keptnv2.ActionTaskName): // sh.keptn.event.action.triggered
		log.Printf("Processing Action.Triggered Event")

		eventData := &keptnv2.ActionTriggeredEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		return HandleActionTriggeredEvent(myKeptn, event, eventData)
	case keptnv2.GetStartedEventType(keptnv2.ActionTaskName): // sh.keptn.event.action.started
		log.Printf("Processing Action.Started Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.ActionStartedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)
	case keptnv2.GetFinishedEventType(keptnv2.ActionTaskName): // sh.keptn.event.action.finished
		log.Printf("Processing Action.Finished Event")
		// Please note: Processing .started, .status.changed and .finished events is only recommended when you want to
		// notify an external service (e.g., for logging purposes).

		eventData := &keptnv2.ActionFinishedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)

	// -------------------------------------------------------
	// sh.keptn.event.problem / problem.open
	// Please Note: This is deprecated; Use Action.Triggered instead
	case keptnlib.ProblemEventType: // sh.keptn.event.problem - e.g., sent by Dynatrace to Keptn Api
		log.Printf("Processing problem Event")
		log.Printf("Subscribing to a problem.open or problem event is not recommended since Keptn 0.7. Please subscribe to event of type: sh.keptn.event.action.triggered")

		eventData := &keptnlib.ProblemEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		return HandleProblemEvent(myKeptn, event, eventData)
	case keptnlib.ProblemOpenEventType: // sh.keptn.event.problem.open - e.g., sent by dynatrace-service
		log.Printf("Processing problem.open Event")
		log.Printf("Subscribing to a problem.open or problem event is not recommended since Keptn 0.7. Please subscribe to event of type: sh.keptn.event.action.triggered")

		eventData := &keptnlib.ProblemEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		return HandleProblemEvent(myKeptn, event, eventData)

	// -------------------------------------------------------
	// sh.keptn.event.get-sli
	case keptnv2.GetTriggeredEventType(keptnv2.GetSLITaskName): // sh.keptn.event.get-sli.triggered
		log.Printf("Processing Get-SLI.Triggered Event")

		eventData := &keptnv2.GetSLITriggeredEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		return HandleGetSliTriggeredEvent(myKeptn, event, eventData)
	case keptnv2.GetStartedEventType(keptnv2.GetSLITaskName): // sh.keptn.event.get-sli.started
		log.Printf("Processing Get-SLI.Started Event")

		eventData := &keptnv2.GetSLIStartedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)
	case keptnv2.GetFinishedEventType(keptnv2.GetSLITaskName): // sh.keptn.event.get-sli.finished
		log.Printf("Processing Get-SLI.Finished Event")

		eventData := &keptnv2.GetSLIFinishedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)

	// -------------------------------------------------------
	// sh.keptn.event.configure-monitoring
	case keptnlib.ConfigureMonitoringEventType: // old configure-monitoring CE; compatibility with 0.8.0-alpha
		log.Printf("Processing old configure-monitoring Event")
		log.Printf("Note: This will be deprecated with Keptn 0.8.0")

		eventData := &keptnlib.ConfigureMonitoringEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Handle old configure-monitoring event
		return OldHandleConfigureMonitoringEvent(myKeptn, event, eventData)

	case keptnv2.GetTriggeredEventType(keptnv2.ConfigureMonitoringTaskName): // sh.keptn.event.configure-monitoring.triggered
		log.Printf("Processing configure-monitoring.Triggered Event")

		eventData := &keptnv2.ConfigureMonitoringTriggeredEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		return HandleConfigureMonitoringTriggeredEvent(myKeptn, event, eventData)
	case keptnv2.GetStartedEventType(keptnv2.ConfigureMonitoringTaskName): // sh.keptn.event.configure-monitoring.started
		log.Printf("Processing configure-monitoring.Started Event")

		eventData := &keptnv2.ConfigureMonitoringStartedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)
	case keptnv2.GetFinishedEventType(keptnv2.ConfigureMonitoringTaskName): // sh.keptn.event.configure-monitoring.finished
		log.Printf("Processing configure-monitoring.Finished Event")

		eventData := &keptnv2.ConfigureMonitoringFinishedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)

	// -------------------------------------------------------
	// your custom cloud event, e.g., sh.keptn.your-event
	// see https://github.com/keptn-sandbox/echo-service/blob/a90207bc119c0aca18368985c7bb80dea47309e9/pkg/events.go
	// for an example on how to generate your own CloudEvents and structs
	case keptnv2.GetTriggeredEventType("your-event"): // sh.keptn.event.your-event.triggered
		log.Printf("Processing your-event.triggered Event")

		// eventData := &keptnv2.YourEventTriggeredEventData{}
		//  parseKeptnCloudEventPayload(event, eventData)

		break
	case keptnv2.GetStartedEventType(keptnv2.ConfigureMonitoringTaskName): // sh.keptn.event.your-event.started
		log.Printf("Processing your-event.started Event")

		// eventData := &keptnv2.YourEventStartedEventData{}
		// parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		// return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)

		break
	case keptnv2.GetFinishedEventType(keptnv2.ConfigureMonitoringTaskName): // sh.keptn.event.your-event.finished
		log.Printf("Processing your-event.finished Event")

		// eventData := &keptnv2.YourEventFinishedEventData{}
		// parseKeptnCloudEventPayload(event, eventData)

		// Just log this event
		// return GenericLogKeptnCloudEventHandler(myKeptn, event, eventData)

		break
	}

	// Unknown Event -> Throw Error!
	var errorMsg string
	errorMsg = fmt.Sprintf("Unhandled Keptn Cloud Event: %s", event.Type())

	log.Print(errorMsg)
	return errors.New(errorMsg)
}

/**
 * Usage: ./main
 * no args: starts listening for cloudnative events on localhost:port/path
 *
 * Environment Variables
 * env=runlocal   -> will fetch resources from local drive instead of configuration service
 */
func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
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

	keptnOptions.ConfigurationServiceURL = env.ConfigurationServiceUrl

	log.Println("Starting keptn-service-template-go...")
	log.Printf("    on Port = %d; Path=%s", env.Port, env.Path)

	ctx := context.Background()
	ctx = cloudevents.WithEncodingStructured(ctx)

	log.Printf("Creating new http handler")

	// configure http server to receive cloudevents
	p, err := cloudevents.NewHTTP(cloudevents.WithPath(env.Path), cloudevents.WithPort(env.Port))

	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("Starting receiver")
	log.Fatal(c.StartReceiver(ctx, processKeptnCloudEvent))

	return 0
}
