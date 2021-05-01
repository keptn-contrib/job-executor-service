package main

import (
	"didiladi/keptn-generic-job-service/pkg/k8s"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"log"
)

/**
* Here are all the handler functions for the individual event
* See https://github.com/keptn/spec/blob/0.8.0-alpha/cloudevents.md for details on the payload
**/

// GenericLogKeptnCloudEventHandler is a generic handler for Keptn Cloud Events that logs the CloudEvent
func GenericLogKeptnCloudEventHandler(myKeptn *keptnv2.Keptn, incomingEvent cloudevents.Event, data interface{}) error {
	log.Printf("Handling %s Event: %s", incomingEvent.Type(), incomingEvent.Context.GetID())
	log.Printf("CloudEvent %T: %v", data, data)

	return nil
}

// HandleActionTriggeredEvent handles action.triggered events
func HandleActionTriggeredEvent(myKeptn *keptnv2.Keptn, incomingEvent cloudevents.Event, data *keptnv2.ActionTriggeredEventData) error {
	log.Printf("Handling Action Triggered Event: %s", incomingEvent.Context.GetID())
	log.Printf("Action=%s\n", data.Action.Action)

	// check if action is supported
	if data.Action.Action == "locust" {

		handleLocust(myKeptn, data)

	} else {
		log.Printf("Retrieved unknown action %s, skipping...", data.Action.Action)
		return nil
	}
	return nil
}

func handleLocust(myKeptn *keptnv2.Keptn, data *keptnv2.ActionTriggeredEventData) {

	myKeptn.SendTaskStartedEvent(data, ServiceName)

	jobName := "test"
	image := "locustio/locust"
	namespace := "keptn"
	cmd := "-f /mnt/locust/locustfile.py"

	pvcName := "locust-test"
	storageClassName := "gp2"

	clientset, err := k8s.ConnectToCluster(namespace)
	if err != nil {
		sendTaskFailedEvent(myKeptn, jobName, err)
		return
	}

	err = k8s.CreateK8sPvc(clientset, namespace, pvcName, storageClassName)
	if err != nil {
		sendTaskFailedEvent(myKeptn, jobName, err)
		return
	}

	err = k8s.CreateK8sJob(clientset, namespace, jobName, image, cmd, pvcName)
	if err != nil {
		sendTaskFailedEvent(myKeptn, jobName, err)
		return
	}

	myKeptn.SendTaskFinishedEvent(&keptnv2.EventData{
		Status:  keptnv2.StatusSucceeded,
		Result:  keptnv2.ResultPass,
		Message: fmt.Sprintf("Job %s finished successfully!", jobName),
	}, ServiceName)
}

func sendTaskFailedEvent(myKeptn *keptnv2.Keptn, jobName string, err error) (string, error) {
	return myKeptn.SendTaskFinishedEvent(&keptnv2.EventData{
		Status:  keptnv2.StatusErrored,
		Result:  keptnv2.ResultFailed,
		Message: fmt.Sprintf("Job %s failed: %s", jobName, err.Error()),
	}, ServiceName)
}
