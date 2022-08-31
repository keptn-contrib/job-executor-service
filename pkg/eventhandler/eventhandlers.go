package eventhandler

import (
	"fmt"
	"github.com/keptn/go-utils/pkg/sdk"
	keptn_interface "keptn-contrib/job-executor-service/pkg/keptn"
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
	Map(event sdk.KeptnEvent) (map[string]interface{}, error)
}

// JobConfigReader retrieves the job-executor-service configuration
type JobConfigReader interface {
	GetJobConfig(gitCommitID string) (*config.Config, string, error)
}

// ErrorLogSender is used to send error logs that will appear in Uniform UI
type ErrorLogSender interface {
	SendErrorLogEvent(initialCloudEvent *cloudevents.Event, applicationError error) error
}

// K8s is used to interact with kubernetes jobs
type K8s interface {
	ConnectToCluster() error
	CreateK8sJob(
		jobName string, jobDetails k8sutils.JobDetails, eventData keptn.EventProperties,
		jobSettings k8sutils.JobSettings, jsonEventData interface{}, namespace string,
	) error
	AwaitK8sJobDone(
		jobName string, maxPollDuration time.Duration, pollIntervalInSeconds time.Duration, namespace string,
	) error
	GetFailedEventsForJob(jobName string, namespace string) (string, error)
	GetLogsOfPod(jobName string, namespace string) (string, error)
}

// EventHandler contains all information needed to process an event
type EventHandler struct {
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

// Execute handles all events in a generic manner
func (eh *EventHandler) Execute(k sdk.IKeptn, event sdk.KeptnEvent) (interface{}, *sdk.Error) {
	eventAsInterface, err := eh.Mapper.Map(event)
	if err != nil {
		log.Printf("failed to convert incoming cloudevent: %v", err)
		return nil, &sdk.Error{Err: err, StatusType: keptnv2.StatusErrored, ResultType: keptnv2.ResultFailed, Message: "failed to convert incoming cloudevent: " + err.Error()}
	}

	k.Logger().Infof(
		"Attempting to handle event %s of type %s ...", event.ID,
		*event.Type,
	)
	k.Logger().Infof("CloudEvent %T: %v", eventAsInterface, eventAsInterface)

	// Get the git commit id from the cloud event (if it exists) and use it to query the job configuration
	var gitCommitID string
	if commitId, ok := eventAsInterface["gitcommitid"]; ok {
		gitCommitID, _ = commitId.(string)
	}

	data := &keptnv2.EventData{}
	if err := keptnv2.Decode(event.Data, data); err != nil {
		k.Logger().Errorf("Could not parse event: %s", err.Error())
		return nil, &sdk.Error{Err: err, StatusType: keptnv2.StatusErrored, ResultType: keptnv2.ResultFailed, Message: fmt.Sprintf("Could not parse event: %s", err.Error())}
	}

	var jobConfigReader JobConfigReader
	if eh.JobConfigReader == nil {
		jobConfigReader = &config.JobConfigReader{
			Keptn: keptn_interface.NewV1ResourceHandler(*data, k.APIV2().Resources()),
		}
	} else {
		// Used to pass a mock in the unit tests
		jobConfigReader = eh.JobConfigReader
	}

	configuration, configHash, err := jobConfigReader.GetJobConfig(gitCommitID)
	if err != nil {
		return nil, &sdk.Error{Err: err, StatusType: keptnv2.StatusErrored, ResultType: keptnv2.ResultFailed, Message: "could not retrieve config for job-executor-service: " + err.Error()}
	}

	// For each action that matches the given event type we execute all containing tasks:
	for actionIndex, action := range configuration.Actions {
		if action.IsEventMatch(*event.Type, eventAsInterface) {
			log.Printf(
				"Match found for event %s of type %s. Starting k8s job to run action '%s'", event.ID,
				*event.Type, action.Name,
			)

			finishedEvent, err := eh.startK8sJob(k, event, data, &action, actionIndex, configHash, gitCommitID, eventAsInterface)
			if err != nil {
				return nil, err
			} else if finishedEvent != nil {
				return finishedEvent, nil
			}
		}
	}

	k.Logger().Info("Returning empty response")
	return nil, nil
}

func (eh *EventHandler) startK8sJob(k sdk.IKeptn, event sdk.KeptnEvent, eventData keptn.EventProperties, action *config.Action, actionIndex int, configHash string, gitCommitID string,
	jsonEventData interface{},
) (interface{}, *sdk.Error) {
	err := eh.K8s.ConnectToCluster()
	if err != nil {
		k.Logger().Infof("Error while connecting to k8s cluster: %e", err)
		if !action.Silent {
			return nil, &sdk.Error{Err: err, StatusType: keptnv2.StatusErrored, ResultType: keptnv2.ResultFailed, Message: fmt.Sprintf("Error while connecting to cluster: %s", err.Error())}
		}
		return nil, nil
	}

	var allJobLogs []jobLogs
	additionalFinishedEventData := dataForFinishedEvent{
		start: time.Now(),
	}

	// To execute all tasks atomically, we check all images
	// before we start executing a single task of a job
	for _, task := range action.Tasks {
		if !eh.ImageFilter.IsImageAllowed(task.Image) {
			errorText := fmt.Sprintf("Forbidden: Image %s does not match configured image allowlist.\n", task.Image)

			k.Logger().Infof(errorText)
			if !action.Silent {
				return nil, &sdk.Error{Err: err, StatusType: keptnv2.StatusErrored, ResultType: keptnv2.ResultFailed, Message: errorText}
			}

			return nil, nil
		}
	}

	for index, task := range action.Tasks {
		k.Logger().Infof("Starting task %s/%s: '%s' ...", strconv.Itoa(index+1), strconv.Itoa(len(action.Tasks)), task.Name)

		// k8s job name max length is 63 characters, with the naming scheme below up to 999 tasks per action are supported
		// the naming scheme is also unique if multiple actions in one cloud event are executed
		jobName := fmt.Sprintf("job-executor-service-job-%s-%03d-%03d", event.ID[:24], actionIndex, index+1)

		namespace := eh.JobSettings.JobNamespace

		if len(task.Namespace) > 0 {
			namespace = task.Namespace
		}

		jobDetails := k8sutils.JobDetails{
			Action:        action,
			Task:          &task,
			ActionIndex:   actionIndex,
			TaskIndex:     index,
			JobConfigHash: configHash,
			GitCommitID:   gitCommitID,
		}

		err = eh.K8s.CreateK8sJob(
			jobName, jobDetails, eventData, eh.JobSettings, jsonEventData, namespace,
		)

		if err != nil {
			k.Logger().Infof("Error while creating job: %s\n", err)
			if !action.Silent {
				return nil, &sdk.Error{Err: err, StatusType: keptnv2.StatusErrored, ResultType: keptnv2.ResultFailed, Message: fmt.Sprintf("Error while creating job: %s", err)}
			}
		}

		maxPollDuration := defaultMaxPollDuration
		if task.MaxPollDuration != nil {
			maxPollDuration = time.Duration(*task.MaxPollDuration) * time.Second
		}
		jobErr := eh.K8s.AwaitK8sJobDone(jobName, maxPollDuration, pollInterval, namespace)

		logs, err := eh.K8s.GetLogsOfPod(jobName, namespace)
		if err != nil {
			k.Logger().Infof("Error while retrieving logs: %s\n", err.Error())
		}

		if jobErr != nil {
			k.Logger().Infof("Error while creating job: %s\n", jobErr.Error())

			erroredEventMessages, eventErr := eh.K8s.GetFailedEventsForJob(jobName, namespace)

			if eventErr != nil {
				k.Logger().Infof("Error while retrieving events: %s\n", eventErr.Error())
			} else if erroredEventMessages != "" {
				// Found some failed events for this job - appending them to logs
				logs = logs + "\n" + erroredEventMessages
			}

			if !action.Silent {
				return nil, &sdk.Error{Err: err, StatusType: keptnv2.StatusErrored, ResultType: keptnv2.ResultFailed, Message: fmt.Sprintf("Error while creating job: %s", jobErr.Error())}
			}
			return nil, nil
		}

		allJobLogs = append(
			allJobLogs, jobLogs{
				name: task.Name,
				logs: logs,
			},
		)
	}

	additionalFinishedEventData.end = time.Now()

	k.Logger().Infof("Successfully finished processing of event: %s\n", event.ID)

	if !action.Silent {
		k.Logger().Infof("Getting task finished event")
		return getTaskFinishedEvent(event, eventData, allJobLogs, additionalFinishedEventData), nil
	}

	return nil, nil
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

// getTaskFinishedEvent returns the finished data for the received event as an interface which can be directly returned using the go-sdk
func getTaskFinishedEvent(event sdk.KeptnEvent, receivedEventData keptn.EventProperties, jobLogs []jobLogs, data dataForFinishedEvent) interface{} {
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
		Project: receivedEventData.GetProject(),
		Stage:   receivedEventData.GetStage(),
		Service: receivedEventData.GetService(),
	}

	if isTestTriggeredEvent(*event.Type) && !data.start.IsZero() && !data.end.IsZero() {
		return keptnv2.TestFinishedEventData{
			Test: keptnv2.TestFinishedDetails{
				Start: data.start.Format(time.RFC3339),
				End:   data.end.Format(time.RFC3339),
			},
			EventData: *eventData,
		}
	} else {
		return eventData
	}
}

func isTestTriggeredEvent(eventName string) bool {
	return eventName == keptnv2.GetTriggeredEventType(keptnv2.TestTaskName)
}
