package eventhandler

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"keptn-contrib/job-executor-service/pkg/config"
	"keptn-contrib/job-executor-service/pkg/k8sutils"

	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

const (
	pollIntervalInSeconds = 5
	defaultMaxPollCount   = 60
)

// EventHandler contains all information needed to process an event
type EventHandler struct {
	Keptn       *keptnv2.Keptn
	Event       cloudevents.Event
	EventData   *keptnv2.EventData
	ServiceName string
	JobSettings k8sutils.JobSettings
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

	eventAsInterface, err := eh.createEventPayloadAsInterface()
	if err != nil {
		log.Printf("failed to convert incoming cloudevent: %v", err)
		return err
	}

	log.Printf("Attempting to handle event %s of type %s ...", eh.Event.Context.GetID(), eh.Event.Type())
	log.Printf("CloudEvent %T: %v", eventAsInterface, eventAsInterface)

	resource, err := eh.Keptn.GetKeptnResource("job/config.yaml")
	if err != nil {
		log.Printf("Could not find config for job-executor-service: %s", err.Error())

		if eh.JobSettings.AlwaysSendFinishedEvent {
			_, err := eh.Keptn.SendTaskStartedEvent(eh.EventData, eh.ServiceName)
			if err != nil {
				log.Printf("Error while sending started event: %s\n", err.Error())
			}
			sendTaskFinishedEvent(eh.Keptn, eh.ServiceName, nil, dataForFinishedEvent{})
		}

		return err
	}

	configuration, err := config.NewConfig(resource)
	if err != nil {
		log.Printf("Could not parse config: %s", err)
		log.Printf("The config was: %s", string(resource))
		return err
	}

	match, action := configuration.IsEventMatch(eh.Event.Type(), eventAsInterface)
	if !match {
		log.Printf("No match found for event %s of type %s. Skipping...", eh.Event.Context.GetID(), eh.Event.Type())
		return nil
	}

	log.Printf("Match found for event %s of type %s. Starting k8s job to run action '%s'", eh.Event.Context.GetID(), eh.Event.Type(), action.Name)

	k8s := k8sutils.NewK8s(eh.JobSettings.JobNamespace)
	eh.startK8sJob(k8s, action, eventAsInterface)

	return nil
}

func (eh *EventHandler) createEventPayloadAsInterface() (map[string]interface{}, error) {

	var eventDataAsInterface interface{}
	err := json.Unmarshal(eh.Event.Data(), &eventDataAsInterface)
	if err != nil {
		log.Printf("failed to convert incoming cloudevent: %v", err)
		return nil, err
	}

	extension, _ := eh.Event.Context.GetExtension("shkeptncontext")
	shKeptnContext := extension.(string)

	eventAsInterface := make(map[string]interface{})
	eventAsInterface["id"] = eh.Event.ID()
	eventAsInterface["shkeptncontext"] = shKeptnContext
	eventAsInterface["time"] = eh.Event.Time()
	eventAsInterface["source"] = eh.Event.Source()
	eventAsInterface["data"] = eventDataAsInterface
	eventAsInterface["specversion"] = eh.Event.SpecVersion()
	eventAsInterface["type"] = eh.Event.Type()

	return eventAsInterface, nil
}

func (eh *EventHandler) startK8sJob(k8s k8sutils.K8s, action *config.Action, jsonEventData interface{}) {

	if !action.Silent {
		_, err := eh.Keptn.SendTaskStartedEvent(eh.EventData, eh.ServiceName)
		if err != nil {
			log.Printf("Error while sending started event: %s\n", err.Error())
			return
		}
	}

	err := k8s.ConnectToCluster()
	if err != nil {
		log.Printf("Error while connecting to cluster: %s\n", err.Error())
		if !action.Silent {
			sendTaskFailedEvent(eh.Keptn, "", eh.ServiceName, err, "")
		}
		return
	}

	var allJobLogs []jobLogs
	additionalFinishedEventData := dataForFinishedEvent{
		start: time.Now(),
	}

	for index, task := range action.Tasks {
		log.Printf("Starting task %s/%s: '%s' ...", strconv.Itoa(index+1), strconv.Itoa(len(action.Tasks)), task.Name)

		// k8s job name max length is 63 characters, with the naming scheme below up to 999 tasks per action are supported
		jobName := "job-executor-service-job-" + eh.Event.ID()[:28] + "-" + strconv.Itoa(index+1)

		namespace := eh.JobSettings.JobNamespace

		if len(task.Namespace) > 0 {
			namespace = task.Namespace
		}

		err := k8s.CreateK8sJob(jobName, action, task, eh.EventData, eh.JobSettings,
			jsonEventData, namespace)

		if err != nil {
			log.Printf("Error while creating job: %s\n", err)
			if !action.Silent {
				sendTaskFailedEvent(eh.Keptn, jobName, eh.ServiceName, err, "")
			}
			return
		}

		maxPollCount := defaultMaxPollCount
		if task.MaxPollDuration != nil {
			maxPollCount = int(math.Ceil(float64(*task.MaxPollDuration) / pollIntervalInSeconds))
		}
		jobErr := k8s.AwaitK8sJobDone(jobName, maxPollCount, pollIntervalInSeconds, namespace)

		logs, err := k8s.GetLogsOfPod(jobName, namespace)
		if err != nil {
			log.Printf("Error while retrieving logs: %s\n", err.Error())
		}

		if jobErr != nil {
			log.Printf("Error while creating job: %s\n", jobErr.Error())
			if !action.Silent {
				sendTaskFailedEvent(eh.Keptn, jobName, eh.ServiceName, jobErr, logs)
			}
			return
		}

		allJobLogs = append(allJobLogs, jobLogs{
			name: jobName,
			logs: logs,
		})
	}

	additionalFinishedEventData.end = time.Now()

	log.Printf("Successfully finished processing of event: %s\n", eh.Event.ID())

	if !action.Silent {
		sendTaskFinishedEvent(eh.Keptn, eh.ServiceName, allJobLogs, additionalFinishedEventData)
	}
}

func sendTaskFailedEvent(myKeptn *keptnv2.Keptn, jobName string, serviceName string, err error, logs string) {
	var message string

	if logs != "" {
		message = fmt.Sprintf("Job %s failed: %s\n\nLogs: \n%s", jobName, err, logs)
	} else {
		message = fmt.Sprintf("Job %s failed: %s", jobName, err)
	}

	_, err = myKeptn.SendTaskFinishedEvent(&keptnv2.EventData{
		Status:  keptnv2.StatusErrored,
		Result:  keptnv2.ResultFailed,
		Message: message,
	}, serviceName)

	if err != nil {
		log.Printf("Error while sending started event: %s\n", err)
	}
}

func sendTaskFinishedEvent(myKeptn *keptnv2.Keptn, serviceName string, jobLogs []jobLogs, data dataForFinishedEvent) {
	var message string

	for _, jobLogs := range jobLogs {
		message += fmt.Sprintf("Job %s finished successfully!\n\nLogs:\n%s\n\n", jobLogs.name, jobLogs.logs)
	}

	eventData := &keptnv2.EventData{

		Status:  keptnv2.StatusSucceeded,
		Result:  keptnv2.ResultPass,
		Message: message,
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
