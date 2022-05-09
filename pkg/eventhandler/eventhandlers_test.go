package eventhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/golang/mock/gomock"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	keptnfake "github.com/keptn/go-utils/pkg/lib/v0_2_0/fake"

	"keptn-contrib/job-executor-service/pkg/config"
	"keptn-contrib/job-executor-service/pkg/k8sutils"

	eventhandlerfake "keptn-contrib/job-executor-service/pkg/eventhandler/fake"
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

var apiVersion = "v2"

type acceptAllImagesFilter struct {
	ImageFilter
}

func (f acceptAllImagesFilter) IsImageAllowed(_ string) bool {
	return true
}

func createK8sMock(t *testing.T) *eventhandlerfake.MockK8s {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	return eventhandlerfake.NewMockK8s(mockCtrl)
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

func TestErrorMappingEvent(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockMapper := eventhandlerfake.NewMockEventMapper(mockCtrl)
	mockFilter := eventhandlerfake.NewMockImageFilter(mockCtrl)
	mockJobConfigReader := eventhandlerfake.NewMockJobConfigReader(mockCtrl)
	mockUniformErrorSender := eventhandlerfake.NewMockErrorLogSender(mockCtrl)

	mappingError := errors.New("some weird error happened during mapping")
	mockMapper.EXPECT().Map(gomock.Any()).Times(1).Return(
		nil,
		mappingError,
	)

	myKeptn, _, _, err := initializeTestObjects("../../test/events/action.triggered.json")

	require.NoError(t, err)

	sut := EventHandler{
		Keptn:           myKeptn,
		JobConfigReader: mockJobConfigReader,
		ServiceName:     "test-jes",
		JobSettings:     k8sutils.JobSettings{},
		ImageFilter:     mockFilter,
		Mapper:          mockMapper,
		ErrorSender:     mockUniformErrorSender,
	}
	err = sut.HandleEvent()
	assert.Error(t, err)
	assert.ErrorIs(t, err, mappingError)
}

type matcherErrorIs struct {
	targetErr error
}

func (mei *matcherErrorIs) Matches(x interface{}) bool {

	actualErr, ok := x.(error)
	if !ok {
		return false
	}

	return errors.Is(actualErr, mei.targetErr)
}

func (mei *matcherErrorIs) String() string {
	return fmt.Sprintf("%#v", mei.targetErr)
}

func TestErrorGettingJobConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFilter := eventhandlerfake.NewMockImageFilter(mockCtrl)
	mockJobConfigReader := eventhandlerfake.NewMockJobConfigReader(mockCtrl)
	mockUniformErrorSender := eventhandlerfake.NewMockErrorLogSender(mockCtrl)

	errorGettingJobConfig := errors.New("error getting resource")
	mockJobConfigReader.EXPECT().GetJobConfig().Return(
		nil,
		errorGettingJobConfig,
	).Times(1)

	myKeptn, _, fakeEventSender, err := initializeTestObjects("../../test/events/action.triggered.json")
	require.NoError(t, err)

	mockUniformErrorSender.EXPECT().SendErrorLogEvent(
		myKeptn.CloudEvent,
		&matcherErrorIs{targetErr: errorGettingJobConfig},
	).Times(1)

	sut := EventHandler{
		Keptn:           myKeptn,
		JobConfigReader: mockJobConfigReader,
		ServiceName:     "",
		JobSettings:     k8sutils.JobSettings{},
		ImageFilter:     mockFilter,
		Mapper:          new(KeptnCloudEventMapper),
		ErrorSender:     mockUniformErrorSender,
	}

	err = sut.HandleEvent()
	assert.ErrorIs(t, err, errorGettingJobConfig)
	assert.NoError(t, fakeEventSender.AssertSentEventTypes([]string{}))
}

func TestErrorConnectingToK8s(t *testing.T) {

	tests := []struct {
		name               string
		silent             bool
		expectedEventTypes []string
	}{
		{
			name:   "Error Connecting with silent=false",
			silent: false,
			expectedEventTypes: []string{
				"sh.keptn.event.action.started",
				"sh.keptn.event.action.finished",
			},
		},
		{
			name:               "Error Connecting with silent=true",
			silent:             true,
			expectedEventTypes: []string{},
		},
	}

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				mockCtrl := gomock.NewController(t)
				defer mockCtrl.Finish()

				mockFilter := eventhandlerfake.NewMockImageFilter(mockCtrl)
				mockJobConfigReader := eventhandlerfake.NewMockJobConfigReader(mockCtrl)
				mockUniformErrorSender := eventhandlerfake.NewMockErrorLogSender(mockCtrl)

				mockJobConfigReader.EXPECT().GetJobConfig().Return(
					&config.Config{
						APIVersion: &apiVersion,
						Actions: []config.Action{
							{
								Name: "catch-all action",
								Events: []config.Event{
									{
										Name:     "sh.keptn.event.*",
										JSONPath: config.JSONPath{},
									},
								},
								Tasks: []config.Task{
									{
										Name:  "sample-task",
										Image: "sampleimage",
										Cmd:   []string{"echo"},
										Args:  []string{"Hello", "World"},
									},
								},
								Silent: test.silent,
							},
						},
					},
					nil,
				).Times(1)

				mockK8s := eventhandlerfake.NewMockK8s(mockCtrl)
				connectionError := errors.New("error connecting to k8s cluster")
				mockK8s.EXPECT().ConnectToCluster().Return(connectionError).Times(1)

				myKeptn, _, mockEventSender, err := initializeTestObjects("../../test/events/action.triggered.json")
				require.NoError(t, err)

				sut := EventHandler{
					Keptn:           myKeptn,
					JobConfigReader: mockJobConfigReader,
					ServiceName:     "test-jes",
					JobSettings:     k8sutils.JobSettings{},
					ImageFilter:     mockFilter,
					Mapper:          new(KeptnCloudEventMapper),
					K8s:             mockK8s,
					ErrorSender:     mockUniformErrorSender,
				}

				err = sut.HandleEvent()

				// TODO: weird that if connection to k8s fails we don't send back an error from the handling function
				assert.NoError(t, err)
				assert.NoError(
					t, mockEventSender.AssertSentEventTypes(
						test.expectedEventTypes,
					),
				)
			},
		)
	}
}

type nameMatcher struct {
	name string
}

func (am *nameMatcher) Matches(x interface{}) bool {
	reflectedValue := reflect.ValueOf(x)
	if reflectedValue.Kind() == reflect.Ptr || reflectedValue.Kind() == reflect.Interface {
		// if it's a pointer or an interface dereference to get to the struct
		reflectedValue = reflectedValue.Elem()
	}

	return am.name == reflectedValue.FieldByName("Name").String()
}

func (am *nameMatcher) String() string {
	return fmt.Sprintf("{Action/Task - name: %s}", am.name)
}

func TestEventMatching(t *testing.T) {

	apiVersion := "apiversion"
	config := config.Config{
		APIVersion: &apiVersion,
		Actions: []config.Action{
			{
				Name: "Action on test.triggered",
				Events: []config.Event{
					{
						Name:     "sh.keptn.event.test.triggered",
						JSONPath: config.JSONPath{},
					},
				},
				Tasks: []config.Task{
					{
						Name:  "task1",
						Files: nil,
						Image: "alpine",
						Cmd:   []string{"echo"},
						Args:  []string{"Hello from task1"},
					},
				},
				Silent: false,
			},
		},
	}

	type jobInvocation struct {
		actionName string
		taskNames  []string
	}

	type testInput struct {
		eventFile string
	}

	type testExpectations struct {
		events          []string
		jobsInvocations []jobInvocation
	}

	tests := []struct {
		name     string
		inputs   testInput
		expected testExpectations
	}{
		{
			name: "Event matching test action",
			inputs: testInput{
				eventFile: "../../test/events/test.triggered.json",
			},
			expected: testExpectations{
				events: []string{"sh.keptn.event.test.started", "sh.keptn.event.test.finished"},
				jobsInvocations: []jobInvocation{
					{
						actionName: config.Actions[0].Name,
						taskNames:  []string{config.Actions[0].Tasks[0].Name},
					},
				},
			},
		},
		{
			name: "No match",
			inputs: testInput{
				eventFile: "../../test/events/action.triggered.json",
			},
			expected: testExpectations{
				events:          []string{},
				jobsInvocations: []jobInvocation{},
			},
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				mockCtrl := gomock.NewController(t)
				defer mockCtrl.Finish()

				mockFilter := eventhandlerfake.NewMockImageFilter(mockCtrl)
				if len(test.expected.jobsInvocations) > 0 {
					mockFilter.EXPECT().IsImageAllowed(gomock.Any()).Return(true).MinTimes(1)
				} else {
					mockFilter.EXPECT().IsImageAllowed(gomock.Any()).Return(true).Times(0)
				}

				mockJobConfigReader := eventhandlerfake.NewMockJobConfigReader(mockCtrl)
				mockJobConfigReader.EXPECT().GetJobConfig().Return(
					&config,
					nil,
				).Times(1)

				mockK8s := eventhandlerfake.NewMockK8s(mockCtrl)
				if len(test.expected.jobsInvocations) > 0 {
					mockK8s.EXPECT().ConnectToCluster().Return(nil).Times(1)
				} else {
					mockK8s.EXPECT().ConnectToCluster().Return(nil).Times(0)
				}

				mockUniformErrorSender := eventhandlerfake.NewMockErrorLogSender(mockCtrl)

				myKeptn, _, mockEventSender, err := initializeTestObjects(test.inputs.eventFile)

				require.NoError(t, err)

				totalNoOfExpectedTasks := 0
				for _, invocation := range test.expected.jobsInvocations {
					for _, taskName := range invocation.taskNames {
						mockK8s.EXPECT().CreateK8sJob(
							gomock.Any(),
							&nameMatcher{
								name: invocation.actionName,
							},
							gomock.Eq("0"),
							&nameMatcher{
								taskName,
							},
							gomock.Eq(myKeptn.Event),
							gomock.Eq(k8sutils.JobSettings{}),
							gomock.Any(),
							gomock.Eq(""),
						).Times(1)
					}
					totalNoOfExpectedTasks += len(invocation.taskNames)
				}

				// NTH: the expectations below could be tailored per task by capturing the jobname in the
				// CreateK8sJob expectation above, to avoid too much complexity we just expect anything for the correct
				// number of times
				mockK8s.EXPECT().AwaitK8sJobDone(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(totalNoOfExpectedTasks)

				mockK8s.EXPECT().GetLogsOfPod(
					gomock.Any(), gomock.Any(),
				).Return(
					"What a wonderful day for Pod, and therefore of course, the world.",
					nil,
				).Times(totalNoOfExpectedTasks)

				sut := EventHandler{
					Keptn:           myKeptn,
					JobConfigReader: mockJobConfigReader,
					ServiceName:     "test-jes",
					JobSettings:     k8sutils.JobSettings{},
					ImageFilter:     mockFilter,
					Mapper:          new(KeptnCloudEventMapper),
					K8s:             mockK8s,
					ErrorSender:     mockUniformErrorSender,
				}

				err = sut.HandleEvent()
				assert.NoError(t, err)
				assert.NoError(t, mockEventSender.AssertSentEventTypes(test.expected.events))
			},
		)
	}
}

func TestStartK8s(t *testing.T) {
	jobNamespace1 := "keptn"
	jobNamespace2 := "keptn-2"
	myKeptn, _, fakeEventSender, err := initializeTestObjects("../../test/events/action.triggered.json")
	require.NoError(t, err)

	k8sMock := createK8sMock(t)

	eventData := &keptnv2.EventData{}
	myKeptn.CloudEvent.DataAs(eventData)
	eh := EventHandler{
		ServiceName: "job-executor-service",
		Keptn:       myKeptn,
		ImageFilter: acceptAllImagesFilter{},
		JobSettings: k8sutils.JobSettings{
			JobNamespace: jobNamespace1,
		},
		K8s: k8sMock,
	}
	mapper := new(KeptnCloudEventMapper)
	eventPayloadAsInterface, err := mapper.Map(*eh.Keptn.CloudEvent)

	maxPollDuration := 1006
	action := config.Action{
		Name: "Run locust",
		Tasks: []config.Task{
			{
				Name:            "Run locust smoked ham tests",
				MaxPollDuration: &maxPollDuration,
			},
			{
				Name:      "Run locust healthy snack tests",
				Namespace: jobNamespace2,
			},
		},
	}

	k8sMock.EXPECT().ConnectToCluster().Times(1)
	k8sMock.EXPECT().CreateK8sJob(
		gomock.Eq(jobName1), gomock.Any(), gomock.Eq("0"), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), jobNamespace1,
	).Times(1)
	k8sMock.EXPECT().CreateK8sJob(
		gomock.Eq(jobName2), gomock.Any(), gomock.Eq("0"), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), jobNamespace2,
	).Times(1)
	k8sMock.EXPECT().AwaitK8sJobDone(gomock.Eq(jobName1), 1006*time.Second, pollInterval, jobNamespace1).Times(1)
	k8sMock.EXPECT().AwaitK8sJobDone(gomock.Eq(jobName2), defaultMaxPollDuration, pollInterval, jobNamespace2).Times(1)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName1), jobNamespace1).Times(1)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName2), jobNamespace2).Times(1)

	eh.startK8sJob(&action, 0, eventPayloadAsInterface)

	err = fakeEventSender.AssertSentEventTypes(
		[]string{
			"sh.keptn.event.action.started", "sh.keptn.event.action.finished",
		},
	)
	assert.NoError(t, err)
}

func TestStartK8sJobSilent(t *testing.T) {
	myKeptn, _, fakeEventSender, err := initializeTestObjects("../../test/events/action.triggered.json")
	require.NoError(t, err)

	k8sMock := createK8sMock(t)

	eventData := &keptnv2.EventData{}
	myKeptn.CloudEvent.DataAs(eventData)
	eh := EventHandler{
		ServiceName: "job-executor-service",
		Keptn:       myKeptn,
		ImageFilter: acceptAllImagesFilter{},
		K8s:         k8sMock,
	}
	mapper := new(KeptnCloudEventMapper)
	eventPayloadAsInterface, err := mapper.Map(*eh.Keptn.CloudEvent)

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

	k8sMock.EXPECT().ConnectToCluster().Times(1)
	k8sMock.EXPECT().CreateK8sJob(
		gomock.Eq(jobName1), gomock.Any(), gomock.Eq("0"), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(),
	).Times(1)
	k8sMock.EXPECT().CreateK8sJob(
		gomock.Eq(jobName2), gomock.Any(), gomock.Eq("0"), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(),
	).Times(1)
	k8sMock.EXPECT().AwaitK8sJobDone(gomock.Any(), defaultMaxPollDuration, pollInterval, "").Times(2)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName1), gomock.Any()).Times(1)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName2), gomock.Any()).Times(1)

	eh.startK8sJob(&action, 0, eventPayloadAsInterface)

	err = fakeEventSender.AssertSentEventTypes([]string{})
	assert.NoError(t, err)
}

func TestStartK8s_TestFinishedEvent(t *testing.T) {
	myKeptn, _, fakeEventSender, err := initializeTestObjects("../../test/events/test.triggered.json")
	require.NoError(t, err)

	k8sMock := createK8sMock(t)

	eventData := &keptnv2.EventData{}
	myKeptn.CloudEvent.DataAs(eventData)
	eh := EventHandler{
		ServiceName: "job-executor-service",
		Keptn:       myKeptn,
		ImageFilter: acceptAllImagesFilter{},
		K8s:         k8sMock,
	}
	mapper := new(KeptnCloudEventMapper)
	eventPayloadAsInterface, err := mapper.Map(*eh.Keptn.CloudEvent)

	action := config.Action{
		Name: "Run locust",
		Tasks: []config.Task{
			{
				Name: "Run locust healthy snack tests",
			},
		},
	}

	k8sMock.EXPECT().ConnectToCluster().Times(1)
	k8sMock.EXPECT().CreateK8sJob(
		gomock.Eq(jobName1), gomock.Any(), gomock.Eq("0"), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(),
	).Times(1)
	k8sMock.EXPECT().AwaitK8sJobDone(gomock.Eq(jobName1), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName1), gomock.Any()).Times(1)

	// set the global timezone for testing
	local, err := time.LoadLocation("UTC")
	require.NoError(t, err)
	time.Local = local

	eh.startK8sJob(&action, 0, eventPayloadAsInterface)

	err = fakeEventSender.AssertSentEventTypes(
		[]string{
			keptnv2.GetStartedEventType(keptnv2.TestTaskName),
			keptnv2.GetFinishedEventType(keptnv2.TestTaskName),
		},
	)
	require.NoError(t, err)

	for _, cloudEvent := range fakeEventSender.SentEvents {
		if cloudEvent.Type() == keptnv2.GetFinishedEventType(keptnv2.TestTaskName) {
			eventData := &keptnv2.TestFinishedEventData{}
			cloudEvent.DataAs(eventData)

			dateLayout := "2006-01-02T15:04:05Z"
			_, err := time.Parse(dateLayout, eventData.Test.Start)
			assert.NoError(t, err)
			_, err = time.Parse(dateLayout, eventData.Test.End)
			assert.NoError(t, err)
		}
	}
}

type disallowAllImagesFilter struct {
	ImageFilter
}

func (f disallowAllImagesFilter) IsImageAllowed(_ string) bool {
	return false
}

func TestExpectImageNotAllowedError(t *testing.T) {
	myKeptn, _, fakeEventSender, err := initializeTestObjects("../../test/events/test.triggered.json")
	require.NoError(t, err)

	k8sMock := createK8sMock(t)

	eventData := &keptnv2.EventData{}
	myKeptn.CloudEvent.DataAs(eventData)
	eh := EventHandler{
		ServiceName: "job-executor-service",
		Keptn:       myKeptn,
		ImageFilter: disallowAllImagesFilter{},
		K8s:         k8sMock,
	}
	mapper := new(KeptnCloudEventMapper)
	eventPayloadAsInterface, err := mapper.Map(*eh.Keptn.CloudEvent)

	notAllowedImageName := "alpine:latest"
	action := config.Action{
		Name: "Run some task with invalid image",
		Tasks: []config.Task{
			{
				Image: notAllowedImageName,
				Name:  "Run some image",
			},
		},
	}

	k8sMock.EXPECT().ConnectToCluster().Times(1)
	k8sMock.EXPECT().CreateK8sJob(
		gomock.Eq(jobName1), gomock.Eq("0"), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(),
	).Times(1)
	k8sMock.EXPECT().AwaitK8sJobDone(gomock.Eq(jobName1), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName1), gomock.Any()).Times(1)

	// set the global timezone for testing
	local, err := time.LoadLocation("UTC")
	require.NoError(t, err)
	time.Local = local

	eh.startK8sJob(&action, 0, eventPayloadAsInterface)

	err = fakeEventSender.AssertSentEventTypes(
		[]string{
			keptnv2.GetStartedEventType(keptnv2.TestTaskName),
			keptnv2.GetFinishedEventType(keptnv2.TestTaskName),
		},
	)
	require.NoError(t, err)

	for _, cloudEvent := range fakeEventSender.SentEvents {
		if cloudEvent.Type() == keptnv2.GetFinishedEventType(keptnv2.TestTaskName) {
			eventData := &keptnv2.TestFinishedEventData{}
			cloudEvent.DataAs(eventData)

			assert.Equal(t, eventData.Status, keptnv2.StatusErrored)
			assert.Equal(t, eventData.Result, keptnv2.ResultFailed)
			assert.Contains(t, eventData.Message, notAllowedImageName)
		}
	}
}
