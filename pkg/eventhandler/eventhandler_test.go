package eventhandler

import (
	"encoding/json"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/golang/mock/gomock"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	keptnfake "github.com/keptn/go-utils/pkg/lib/v0_2_0/fake"
	"gotest.tools/assert"
	"io/ioutil"
	"keptn-sandbox/job-executor-service/pkg/config"
	k8sutilsfake "keptn-sandbox/job-executor-service/pkg/k8sutils/fake"
	"testing"
	"time"
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

const jobName1 = "job-executor-service-job-f2b878d3-03c0-4e8f-bc3f-454b-1"
const jobName2 = "job-executor-service-job-f2b878d3-03c0-4e8f-bc3f-454b-2"

func createK8sMock(t *testing.T) *k8sutilsfake.MockK8s {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	return k8sutilsfake.NewMockK8s(mockCtrl)
}

/**
 * loads a cloud event from the passed test json file and initializes a keptn object with it
 */
func initializeTestObjects(eventFileName string) (*keptnv2.Keptn, *cloudevents.Event, *keptnfake.EventSender, error) {
	// load sample event
	eventFile, err := ioutil.ReadFile(eventFileName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Cant load %s: %s", eventFileName, err.Error())
	}

	incomingEvent := &cloudevents.Event{}

	err = json.Unmarshal(eventFile, incomingEvent)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error parsing: %s", err.Error())
	}

	// Add a Fake EventSender to KeptnOptions
	fakeEventSender := &keptnfake.EventSender{}
	var keptnOptions = keptn.KeptnOpts{
		EventSender: fakeEventSender,
	}
	keptnOptions.UseLocalFileSystem = true
	myKeptn, err := keptnv2.NewKeptn(incomingEvent, keptnOptions)

	return myKeptn, incomingEvent, fakeEventSender, err
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

func TestStartK8s(t *testing.T) {
	myKeptn, event, fakeEventSender, err := initializeTestObjects("../../test-events/action.triggered.json")
	assert.NilError(t, err)

	eventData := &keptnv2.EventData{}
	myKeptn.CloudEvent.DataAs(eventData)
	eh := EventHandler{
		ServiceName: "job-executor-service",
		Keptn:       myKeptn,
		EventData:   eventData,
		Event:       *event,
	}
	eventPayloadAsInterface, _ := eh.createEventPayloadAsInterface()

	action := config.Action{
		Name: "Run locust",
		Tasks: []config.Task{
			{
				Name: "Run locust smoked ham tests",
			},
			{
				Name: "Run locust healthy snack tests",
			},
		},
	}

	k8sMock := createK8sMock(t)
	k8sMock.EXPECT().ConnectToCluster().Times(1)
	k8sMock.EXPECT().CreateK8sJob(gomock.Eq(jobName1), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any()).Return(jobName1, nil).Times(1)
	k8sMock.EXPECT().CreateK8sJob(gomock.Eq(jobName2), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any()).Return(jobName2, nil).Times(1)
	k8sMock.EXPECT().AwaitK8sJobDone(gomock.Any()).Times(2)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName1)).Times(1)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName2)).Times(1)
	k8sMock.EXPECT().DeleteK8sJob(gomock.Eq(jobName1)).Times(1)
	k8sMock.EXPECT().DeleteK8sJob(gomock.Eq(jobName2)).Times(1)

	eh.startK8sJob(k8sMock, &action, eventPayloadAsInterface)

	err = fakeEventSender.AssertSentEventTypes([]string{"sh.keptn.event.action.started", "sh.keptn.event.action.finished"})
	assert.NilError(t, err)
}

func TestStartK8sJobSilent(t *testing.T) {
	myKeptn, event, fakeEventSender, err := initializeTestObjects("../../test-events/action.triggered.json")
	assert.NilError(t, err)

	eventData := &keptnv2.EventData{}
	myKeptn.CloudEvent.DataAs(eventData)
	eh := EventHandler{
		ServiceName: "job-executor-service",
		Keptn:       myKeptn,
		EventData:   eventData,
		Event:       *event,
	}
	eventPayloadAsInterface, _ := eh.createEventPayloadAsInterface()

	action := config.Action{
		Name: "Run locust",
		Tasks: []config.Task{
			{
				Name: "Run locust smoked ham tests",
			},
			{
				Name: "Run locust healthy snack tests",
			},
		},
		Silent: true,
	}

	k8sMock := createK8sMock(t)
	k8sMock.EXPECT().ConnectToCluster().Times(1)
	k8sMock.EXPECT().CreateK8sJob(gomock.Eq(jobName1), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any()).Return(jobName1, nil).Times(1)
	k8sMock.EXPECT().CreateK8sJob(gomock.Eq(jobName2), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any()).Return(jobName2, nil).Times(1)
	k8sMock.EXPECT().AwaitK8sJobDone(gomock.Any()).Times(2)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName1)).Times(1)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName2)).Times(1)
	k8sMock.EXPECT().DeleteK8sJob(gomock.Eq(jobName1)).Times(1)
	k8sMock.EXPECT().DeleteK8sJob(gomock.Eq(jobName2)).Times(1)

	eh.startK8sJob(k8sMock, &action, eventPayloadAsInterface)

	err = fakeEventSender.AssertSentEventTypes([]string{})
	assert.NilError(t, err)
}
