package eventhandler

import (
	"didiladi/keptn-generic-job-service/pkg/config"
	"didiladi/keptn-generic-job-service/pkg/k8s"
	"fmt"
	"log"
	"os"
	"strconv"

	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

/**
* Here are all the handler functions for the individual event
* See https://github.com/keptn/spec/blob/0.8.0-alpha/cloudevents.md for details on the payload
**/

// HandleEvent handles all events in a generic manner
func HandleEvent(myKeptn *keptnv2.Keptn, incomingEvent cloudevents.Event, data interface{}, eventData *keptnv2.EventData, serviceName string) error {

	log.Printf("Attempting to handle event %s of type %s ...", incomingEvent.Context.GetID(), incomingEvent.Type())
	log.Printf("CloudEvent %T: %v", data, data)

	resource, err := myKeptn.GetKeptnResource("generic-job/config.yaml")
	if err != nil {
		log.Printf("Could not find config for generic Job service")
		return err
	}

	/* For testing:

	configuration, err := config.NewConfig(resource)
	const complexConfig = `
actions:
  - name: "Run locust"
    event: "sh.keptn.event.test.triggered"
    jsonpath:
      property: "$.test.teststrategy" 
      match: "locust"
    tasks:
      - name: "Run locust smoke tests"
        files: 
          - locust/basic.py
          - locust/import.py
        image: "locustio/locust"
        cmd: "locust -f /keptn/locust/locustfile.py"

  - name: "Run bash"
    event: "sh.keptn.event.action.triggered"
    jsonpath: 
      property: "$.action.action"
      match: "hello"
    tasks:
      - name: "Run static world"
        image: "bash"
        cmd: "echo static"
      - name: "Run hello world"
        files: 
          - hello/hello-world.txt
        image: "bash"
        cmd: "cat /keptn/hello/heppo-world.txt | echo"
`
	configuration, err := config.NewConfig([]byte(complexConfig))
	*/

	configuration, err := config.NewConfig(resource)
	if err != nil {
		log.Printf("Could not parse config: %s", err)
		return err
	}

	match, action := configuration.IsEventMatch(incomingEvent.Type(), data)
	if !match {
		log.Printf("No match found for event %s of type %s. Skipping...", incomingEvent.Context.GetID(), incomingEvent.Type())
		return nil
	}

	log.Printf("Match found for event %s of type %s. Starting k8s job to run action '%s'", incomingEvent.Context.GetID(), incomingEvent.Type(), action.Name)

	startK8sJob(myKeptn, eventData, action, serviceName)

	return nil
}

func startK8sJob(myKeptn *keptnv2.Keptn, eventData *keptnv2.EventData, action *config.Action, serviceName string) {

	event, err := myKeptn.SendTaskStartedEvent(eventData, serviceName)
	if err != nil {
		log.Printf("Error while sending started event: %s\n", err.Error())
		return
	}

	namespace, _ := os.LookupEnv("JOB_NAMESPACE")

	for index, task := range action.Tasks {
		log.Printf("Starting task %s/%s: '%s' ...", strconv.Itoa(index + 1), strconv.Itoa(len(action.Tasks)), task.Name)

		jobName := "keptn-generic-job-" + event + "-" + strconv.Itoa(index + 1)

		clientset, err := k8s.ConnectToCluster(namespace)
		if err != nil {
			log.Printf("Error while connecting to cluster: %s\n", err.Error())
			sendTaskFailedEvent(myKeptn, jobName, serviceName, err)
			return
		}

		actionName := action.Name
		configurationServiceUrl := myKeptn.ResourceHandler.BaseURL

		err = k8s.CreateK8sJob(clientset, namespace, jobName, task, eventData, actionName, configurationServiceUrl)
		defer func() {
			err = k8s.DeleteK8sJob(clientset, namespace, jobName)
			if err != nil {
				log.Printf("Error while deleting job: %s\n", err.Error())
			}
		}()

		// TODO get the logs of the job

		if err != nil {
			log.Printf("Error while creating job: %s\n", err.Error())
			sendTaskFailedEvent(myKeptn, jobName, serviceName, err)
			return
		}
	}

	log.Printf("Successfully finished processing of event: %s\n", myKeptn.CloudEvent.ID())

	myKeptn.SendTaskFinishedEvent(&keptnv2.EventData{
		Status:  keptnv2.StatusSucceeded,
		Result:  keptnv2.ResultPass,
		Message: fmt.Sprintf("Job %s finished successfully!", "keptn-generic-job-" + event),
	}, serviceName)
}

func sendTaskFailedEvent(myKeptn *keptnv2.Keptn, jobName string, serviceName string, err error) {

	_, err = myKeptn.SendTaskFinishedEvent(&keptnv2.EventData{
		Status:  keptnv2.StatusErrored,
		Result:  keptnv2.ResultFailed,
		Message: fmt.Sprintf("Job %s failed: %s", jobName, err.Error()),
	}, serviceName)

	if err != nil {
		log.Printf("Error while sending started event: %s\n", err.Error())
	}
}
