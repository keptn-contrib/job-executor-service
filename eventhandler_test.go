package main

import (
	"encoding/json"
	"fmt"
	"github.com/keptn/go-utils/pkg/lib/v0_2_0/fake"
	"io/ioutil"
	"testing"

	keptn "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
)

/**
 * loads a cloud event from the passed test json file and initializes a keptn object with it
 */
func initializeTestObjects(eventFileName string) (*keptnv2.Keptn, *cloudevents.Event, error) {
	// load sample event
	eventFile, err := ioutil.ReadFile(eventFileName)
	if err != nil {
		return nil, nil, fmt.Errorf("Cant load %s: %s", eventFileName, err.Error())
	}

	incomingEvent := &cloudevents.Event{}
	err = json.Unmarshal(eventFile, incomingEvent)
	if err != nil {
		return nil, nil, fmt.Errorf("Error parsing: %s", err.Error())
	}

	// Add a Fake EventSender to KeptnOptions
	var keptnOptions = keptn.KeptnOpts{
		EventSender: &fake.EventSender{},
	}
	keptnOptions.UseLocalFileSystem = true
	myKeptn, err := keptnv2.NewKeptn(incomingEvent, keptnOptions)

	return myKeptn, incomingEvent, err
}

// Tests HandleActionTriggeredEvent
// TODO: Add your test-code
func TestHandleActionTriggeredEvent(t *testing.T) {
	myKeptn, incomingEvent, err := initializeTestObjects("test-events/action.triggered.json")
	if err != nil {
		t.Error(err)
		return
	}

	specificEvent := &keptnv2.ActionTriggeredEventData{}
	err = incomingEvent.DataAs(specificEvent)
	if err != nil {
		t.Errorf("Error getting keptn event data")
	}

	err = HandleActionTriggeredEvent(myKeptn, *incomingEvent, specificEvent)
	if err != nil {
		t.Errorf("Error: " + err.Error())
	}

	gotEvents := len(myKeptn.EventSender.(*fake.EventSender).SentEvents)

	// Verify that HandleGetSliTriggeredEvent has sent 2 cloudevents
	if gotEvents != 2 {
		t.Errorf("Expected two events to be sent, but got %v", gotEvents)
	}

	// Verify that the first CE sent is a .started event
	if keptnv2.GetStartedEventType(keptnv2.ActionTaskName) != myKeptn.EventSender.(*fake.EventSender).SentEvents[0].Type() {
		t.Errorf("Expected a action.started event type")
	}

	// Verify that the second CE sent is a .finished event
	if keptnv2.GetFinishedEventType(keptnv2.ActionTaskName) != myKeptn.EventSender.(*fake.EventSender).SentEvents[1].Type() {
		t.Errorf("Expected a action.finished event type")
	}
}

// Tests HandleDeploymentTriggeredEvent
// TODO: Add your test-code
func TestHandleDeploymentTriggeredEvent(t *testing.T) {
	myKeptn, incomingEvent, err := initializeTestObjects("test-events/evaluation.triggered.json")
	if err != nil {
		t.Error(err)
		return
	}

	specificEvent := &keptnv2.DeploymentTriggeredEventData{}
	err = incomingEvent.DataAs(specificEvent)
	if err != nil {
		t.Errorf("Error getting keptn event data")
	}

	err = HandleDeploymentTriggeredEvent(myKeptn, *incomingEvent, specificEvent)
	if err != nil {
		t.Errorf("Error: " + err.Error())
	}
}

// Tests HandleEvaluationTriggeredEvent
// TODO: Add your test-code
func TestHandleEvaluationTriggeredEvent(t *testing.T) {
	myKeptn, incomingEvent, err := initializeTestObjects("test-events/evaluation.triggered.json")
	if err != nil {
		t.Error(err)
		return
	}

	specificEvent := &keptnv2.EvaluationTriggeredEventData{}
	err = incomingEvent.DataAs(specificEvent)
	if err != nil {
		t.Errorf("Error getting keptn event data")
	}

	err = HandleEvaluationTriggeredEvent(myKeptn, *incomingEvent, specificEvent)
	if err != nil {
		t.Errorf("Error: " + err.Error())
	}
}

// Tests the HandleGetSliTriggeredEvent Handler
// TODO: Add your test-code
func TestHandleGetSliTriggered(t *testing.T) {
	myKeptn, incomingEvent, err := initializeTestObjects("test-events/get-sli.triggered.json")
	if err != nil {
		t.Error(err)
		return
	}

	specificEvent := &keptnv2.GetSLITriggeredEventData{}
	err = incomingEvent.DataAs(specificEvent)
	if err != nil {
		t.Errorf("Error getting keptn event data")
	}

	err = HandleGetSliTriggeredEvent(myKeptn, *incomingEvent, specificEvent)
	if err != nil {
		t.Errorf("Error: " + err.Error())
	}

	gotEvents := len(myKeptn.EventSender.(*fake.EventSender).SentEvents)

	// Verify that HandleGetSliTriggeredEvent has sent 2 cloudevents
	if gotEvents != 2 {
		t.Errorf("Expected two events to be sent, but got %v", gotEvents)
	}

	// Verify that the first CE sent is a .started event
	if keptnv2.GetStartedEventType(keptnv2.GetSLITaskName) != myKeptn.EventSender.(*fake.EventSender).SentEvents[0].Type() {
		t.Errorf("Expected a get-sli.started event type")
	}

	// Verify that the second CE sent is a .finished event
	if keptnv2.GetFinishedEventType(keptnv2.GetSLITaskName) != myKeptn.EventSender.(*fake.EventSender).SentEvents[1].Type() {
		t.Errorf("Expected a get-sli.finished event type")
	}
}

// Tests the HandleReleaseTriggeredEvent Handler
// TODO: Add your test-code
func TestHandleReleaseTriggeredEvent(t *testing.T) {
	myKeptn, incomingEvent, err := initializeTestObjects("test-events/release.triggered.json")
	if err != nil {
		t.Error(err)
		return
	}

	specificEvent := &keptnv2.ReleaseTriggeredEventData{}
	err = incomingEvent.DataAs(specificEvent)
	if err != nil {
		t.Errorf("Error getting keptn event data")
	}

	err = HandleReleaseTriggeredEvent(myKeptn, *incomingEvent, specificEvent)
	if err != nil {
		t.Errorf("Error: " + err.Error())
	}
}
