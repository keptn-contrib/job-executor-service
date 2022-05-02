package eventhandler

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/keptn/go-utils/pkg/lib/keptn"

	"keptn-contrib/job-executor-service/pkg/config"
	"keptn-contrib/job-executor-service/pkg/k8sutils"

	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

const (
	pollInterval           = 5 * time.Second
	defaultMaxPollDuration = 5 * time.Minute
)

//go:generate mockgen -destination=fake/eventhandlers_mock.go -package=fake .  ImageFilter,EventMapper,JobConfigReader,K8s,ErrorLogSender

// ImageFilter provides an interface for the EventHandler to check if an image is allowed to be used in the job tasks
type ImageFilter interface {
	IsImageAllowed(image string) bool
}

// EventMapper represents an object able to map the cloudevent (including its data) into a map that will contain the
// parsed JSON of the event data
type EventMapper interface {
	// Map transforms a cloud event into a generic map[string]interface{}
	Map(cloudevents.Event) (map[string]interface{}, error)
}

// JobConfigReader retrieves the job-executor-service configuration
type JobConfigReader interface {
	GetJobConfig() (*config.Config, error)
}

// ErrorLogSender is used to send error logs that will appear in Uniform UI
type ErrorLogSender interface {
	SendErrorLogEvent(initialCloudEvent *cloudevents.Event, applicationError error) error
}

// K8s is used to interact with kubernetes jobs
type K8s interface {
	ConnectToCluster() error
	CreateK8sJob(
		jobName string, action *config.Action, task config.Task, eventData keptn.EventProperties,
		jobSettings k8sutils.JobSettings, jsonEventData interface{}, namespace string,
	) error
	AwaitK8sJobDone(
		jobName string, maxPollDuration time.Duration, pollIntervalInSeconds time.Duration, namespace string,
	) error
	GetLogsOfPod(jobName string, namespace string) (string, error)
	ExistsServiceAccount(saName string, namespace string) bool
}

// EventHandler contains all information needed to process an event
type EventHandler struct {
	Keptn           *keptnv2.Keptn
	ServiceName     string
	JobConfigReader JobConfigReader
	JobSettings     k8sutils.JobSettings
	ImageFilter     ImageFilter
	Mapper          EventMapper
	K8s             K8s
	ErrorSender     ErrorLogSender
}

type jobLogs struct {
	name string
	logs string
}

type dataForFinishedEvent struct {
	start time.Time
	end   time.Time
}

// HandleEvent handles all events in a generic manner
func (eh *EventHandler) HandleEvent() error {

	eventAsInterface, err := eh.Mapper.Map(*eh.Keptn.CloudEvent)
	if err != nil {
		log.Printf("failed to convert incoming cloudevent: %v", err)
		return err
	}

	log.Printf(
		"Attempting to handle event %s of type %s ...", eh.Keptn.CloudEvent.Context.GetID(),
		eh.Keptn.CloudEvent.Type(),
	)
	log.Printf("CloudEvent %T: %v", eventAsInterface, eventAsInterface)

	configuration, err := eh.JobConfigReader.GetJobConfig()

	if err != nil {
		errorLogErr := eh.ErrorSender.SendErrorLogEvent(
			eh.Keptn.CloudEvent, fmt.Errorf(
				"could not retrieve config for job-executor-service: %w", err,
			),
		)

		if errorLogErr != nil {
			log.Printf(
				"Failed sending error log for keptn context %s: %v. Initial error: %v", eh.Keptn.KeptnContext,
				errorLogErr, err,
			)
		}

		return err
	}

	match, action := configuration.IsEventMatch(eh.Keptn.CloudEvent.Type(), eventAsInterface)
	if !match {
		log.Printf(
			"No match found for event %s of type %s. Skipping...", eh.Keptn.CloudEvent.Context.GetID(),
			eh.Keptn.CloudEvent.Type(),
		)
		return nil
	}

	log.Printf(
		"Match found for event %s of type %s. Starting k8s job to run action '%s'", eh.Keptn.CloudEvent.Context.GetID(),
		eh.Keptn.CloudEvent.Type(), action.Name,
	)

	eh.startK8sJob(action, eventAsInterface)

	return nil
}

func (eh *EventHandler) startK8sJob(action *config.Action, jsonEventData interface{}) {

	if !action.Silent {
		_, err := eh.Keptn.SendTaskStartedEvent(nil, eh.ServiceName)
		if err != nil {
			log.Printf("Error while sending started event: %s\n", err.Error())
			return
		}
	}

	err := eh.K8s.ConnectToCluster()
	if err != nil {
		log.Printf("Error while connecting to cluster: %s\n", err.Error())
		if !action.Silent {
			sendJobFailedEvent(eh.Keptn, "", eh.ServiceName, err)
		}
		return
	}

	var allJobLogs []jobLogs
	additionalFinishedEventData := dataForFinishedEvent{
		start: time.Now(),
	}

	// To execute all tasks atomically, we check all images before we start executing a single task of a job
	// Additionally we want to check if the job configuration is sound (like validating the specified serviceAccounts)
	for _, task := range action.Tasks {

		namespace := eh.JobSettings.JobNamespace
		if len(task.Namespace) > 0 {
			namespace = task.Namespace
		}

		if !eh.ImageFilter.IsImageAllowed(task.Image) {
			errorText := fmt.Sprintf("Forbidden: Image %s does not match configured image allowlist.\n", task.Image)

			log.Printf(errorText)
			if !action.Silent {
				sendTaskFailedEvent(eh.Keptn, task.Name, eh.ServiceName, errors.New(errorText), "")
			}

			return
		}

		if task.ServiceAccount != nil && !eh.K8s.ExistsServiceAccount(*task.ServiceAccount, namespace) {
			errorText := fmt.Sprintf("Error: service account %s does not exist!\n", *task.ServiceAccount)

			log.Printf(errorText)
			if !action.Silent {
				sendTaskFailedEvent(eh.Keptn, task.Name, eh.ServiceName, errors.New(errorText), "")
			}

			return
		}
	}

	for index, task := range action.Tasks {
		log.Printf("Starting task %s/%s: '%s' ...", strconv.Itoa(index+1), strconv.Itoa(len(action.Tasks)), task.Name)

		// k8s job name max length is 63 characters, with the naming scheme below up to 999 tasks per action are supported
		jobName := "job-executor-service-job-" + eh.Keptn.CloudEvent.ID()[:28] + "-" + strconv.Itoa(index+1)

		namespace := eh.JobSettings.JobNamespace

		if len(task.Namespace) > 0 {
			namespace = task.Namespace
		}

		err := eh.K8s.CreateK8sJob(
			jobName, action, task, eh.Keptn.Event, eh.JobSettings,
			jsonEventData, namespace,
		)

		if err != nil {
			log.Printf("Error while creating job: %s\n", err)
			if !action.Silent {
				sendTaskFailedEvent(eh.Keptn, task.Name, eh.ServiceName, err, "")
			}
			return
		}

		maxPollDuration := defaultMaxPollDuration
		if task.MaxPollDuration != nil {
			maxPollDuration = time.Duration(*task.MaxPollDuration) * time.Second
		}
		jobErr := eh.K8s.AwaitK8sJobDone(jobName, maxPollDuration, pollInterval, namespace)

		logs, err := eh.K8s.GetLogsOfPod(jobName, namespace)
		if err != nil {
			log.Printf("Error while retrieving logs: %s\n", err.Error())
		}

		if jobErr != nil {
			log.Printf("Error while creating job: %s\n", jobErr.Error())
			if !action.Silent {
				sendTaskFailedEvent(eh.Keptn, task.Name, eh.ServiceName, jobErr, logs)
			}
			return
		}

		allJobLogs = append(
			allJobLogs, jobLogs{
				name: task.Name,
				logs: logs,
			},
		)
	}

	additionalFinishedEventData.end = time.Now()

	log.Printf("Successfully finished processing of event: %s\n", eh.Keptn.CloudEvent.ID())

	if !action.Silent {
		sendTaskFinishedEvent(eh.Keptn, eh.ServiceName, allJobLogs, additionalFinishedEventData)
	}
}

func sendTaskFailedEvent(myKeptn *keptnv2.Keptn, taskName string, serviceName string, err error, logs string) {
	var message string

	if logs != "" {
		message = fmt.Sprintf("Task '%s' failed: %s\n\nLogs: \n%s", taskName, err.Error(), logs)
	} else {
		message = fmt.Sprintf("Task '%s' failed: %s", taskName, err.Error())
	}

	_, err = myKeptn.SendTaskFinishedEvent(
		&keptnv2.EventData{
			Status:  keptnv2.StatusErrored,
			Result:  keptnv2.ResultFailed,
			Message: message,
		}, serviceName,
	)

	if err != nil {
		log.Printf("Error while sending started event: %s\n", err)
	}
}

func sendJobFailedEvent(myKeptn *keptnv2.Keptn, jobName string, serviceName string, err error) {
	_, err = myKeptn.SendTaskFinishedEvent(
		&keptnv2.EventData{
			Status:  keptnv2.StatusErrored,
			Result:  keptnv2.ResultFailed,
			Message: fmt.Sprintf("Job %s failed: %s", jobName, err.Error()),
		}, serviceName,
	)

	if err != nil {
		log.Printf("Error while sending started event: %s\n", err)
	}
}

func sendTaskFinishedEvent(myKeptn *keptnv2.Keptn, serviceName string, jobLogs []jobLogs, data dataForFinishedEvent) {
	var logMessage strings.Builder

	for _, jobLogs := range jobLogs {
		logMessage.WriteString(
			fmt.Sprintf("Task '%s' finished successfully!\n\nLogs:\n%s\n\n", jobLogs.name, jobLogs.logs),
		)
	}

	eventData := &keptnv2.EventData{
		Status:  keptnv2.StatusSucceeded,
		Result:  keptnv2.ResultPass,
		Message: logMessage.String(),
	}

	var err error

	if isTestTriggeredEvent(myKeptn.CloudEvent.Type()) && !data.start.IsZero() && !data.end.IsZero() {
		event := &keptnv2.TestFinishedEventData{
			Test: keptnv2.TestFinishedDetails{
				Start: data.start.Format(time.RFC3339),
				End:   data.end.Format(time.RFC3339),
			},
			EventData: *eventData,
		}
		_, err = myKeptn.SendTaskFinishedEvent(event, serviceName)
	} else {
		_, err = myKeptn.SendTaskFinishedEvent(eventData, serviceName)
	}

	if err != nil {
		log.Printf("Error while sending finished event: %s\n", err.Error())
		return
	}
}

func isTestTriggeredEvent(eventName string) bool {
	return eventName == keptnv2.GetTriggeredEventType(keptnv2.TestTaskName)
}
