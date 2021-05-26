package eventhandler

import (
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/keptn/go-utils/pkg/lib/v0_2_0/fake"
	"gotest.tools/assert"
	"io/ioutil"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
)

const testEvent = `
{
      "project": "sockshop",
      "stage": "dev",
      "service": "carts",
      "labels": {
        "testId": "4711",
        "buildId": "build-17",
        "owner": "JohnDoe"
      },
      "status": "succeeded",
      "result": "pass",
      "action": {
        "name": "run locust tests",
        "action": "hello",
        "description": "so something as defined in remediation.yaml",
        "value" : "1"
      }
}`

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

func TestInitializeEventPayloadAsInterface(t *testing.T) {

	context := spec.V1.NewContext()
	context.SetID("0123")
	context.SetSource("sourcysource")
	now := time.Now()
	context.SetTime(now)
	context.SetExtension("shkeptncontext", interface{}("mycontext"))

	eh := EventHandler{
		Event: event.Event{
			Context:     context,
			DataEncoded: []byte(testEvent),
		},
	}

	eventPayloadAsInterface, err := eh.createEventPayloadAsInterface()
	assert.NilError(t, err)

	assert.Equal(t, eventPayloadAsInterface["id"], "0123")
	assert.Equal(t, eventPayloadAsInterface["source"], "sourcysource")
	assert.Equal(t, eventPayloadAsInterface["time"], now)
	assert.Equal(t, eventPayloadAsInterface["shkeptncontext"], "mycontext")

	data := eventPayloadAsInterface["data"]
	dataAsMap := data.(map[string]interface{})

	assert.Equal(t, dataAsMap["project"], "sockshop")
}
