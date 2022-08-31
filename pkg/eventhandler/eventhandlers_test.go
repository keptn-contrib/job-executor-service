package eventhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/golang/mock/gomock"
	"github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/models"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	keptnfake "github.com/keptn/go-utils/pkg/lib/v0_2_0/fake"
	"github.com/keptn/go-utils/pkg/sdk"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"keptn-contrib/job-executor-service/pkg/config"
	eventhandlerfake "keptn-contrib/job-executor-service/pkg/eventhandler/fake"
	"keptn-contrib/job-executor-service/pkg/k8sutils"
	"log"
	"reflect"
	"strings"
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

const jobName1 = "job-executor-service-job-f2b878d3-03c0-4e8f-bc3f--000-001"
const jobName2 = "job-executor-service-job-f2b878d3-03c0-4e8f-bc3f--000-002"

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
	myKeptn, err := keptnv2.NewKeptn(incomingEvent, keptnOptions)

	return myKeptn, incomingEvent, fakeEventSender, err
}

func newEvent(filename string) keptnapi.KeptnContextExtendedCE {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	event := keptnapi.KeptnContextExtendedCE{}
	err = json.Unmarshal(content, &event)
	_ = err
	return event
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

	sut := EventHandler{
		JobConfigReader: mockJobConfigReader,
		ServiceName:     "test-jes",
		JobSettings:     k8sutils.JobSettings{},
		ImageFilter:     mockFilter,
		Mapper:          mockMapper,
		ErrorSender:     mockUniformErrorSender,
	}
	fakeKeptn := sdk.NewFakeKeptn("test-job-executor-service")
	fakeKeptn.AddTaskHandler("*", &sut)

	err := fakeKeptn.NewEvent(newEvent("../../test/events/action.triggered.json"))

	require.NoError(t, err)

	fakeKeptn.AssertSentEventStatus(t, 1, keptnv2.StatusErrored)
	fakeKeptn.AssertSentEventResult(t, 1, keptnv2.ResultFailed)
	fakeKeptn.AssertSentEvent(t, 1, func(ce models.KeptnContextExtendedCE) bool {
		getActionFinishedData := keptnv2.GetActionFinishedEventData{}
		ce.DataAs(&getActionFinishedData)
		return strings.Contains(getActionFinishedData.Message, mappingError.Error())
	})
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

	mockJobConfigReader.EXPECT().GetJobConfig("").Return(
		nil, "",
		errorGettingJobConfig,
	).Times(1)

	sut := EventHandler{
		JobConfigReader: mockJobConfigReader,
		ServiceName:     "",
		JobSettings:     k8sutils.JobSettings{},
		ImageFilter:     mockFilter,
		Mapper:          new(KeptnCloudEventMapper),
		ErrorSender:     mockUniformErrorSender,
	}

	fakeKeptn := sdk.NewFakeKeptn("test-job-executor-service")
	fakeKeptn.AddTaskHandler("*", &sut)

	err := fakeKeptn.NewEvent(newEvent("../../test/events/action.triggered.json"))

	require.NoError(t, err)

	fakeKeptn.AssertSentEventType(t, 1, "sh.keptn.event.action.finished")
	fakeKeptn.AssertSentEventStatus(t, 1, keptnv2.StatusErrored)
	fakeKeptn.AssertSentEventResult(t, 1, keptnv2.ResultFailed)
	fakeKeptn.AssertSentEvent(t, 1, func(ce models.KeptnContextExtendedCE) bool {
		getActionFinishedData := keptnv2.GetActionFinishedEventData{}
		ce.DataAs(&getActionFinishedData)
		return strings.Contains(getActionFinishedData.Message, errorGettingJobConfig.Error())
	})
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

				mockJobConfigReader.EXPECT().GetJobConfig("").Return(
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
					}, "",
					nil,
				).Times(1)

				mockK8s := eventhandlerfake.NewMockK8s(mockCtrl)
				connectionError := errors.New("error connecting to k8s cluster")
				mockK8s.EXPECT().ConnectToCluster().Return(connectionError).Times(1)

				sut := EventHandler{
					JobConfigReader: mockJobConfigReader,
					ServiceName:     "test-jes",
					JobSettings:     k8sutils.JobSettings{},
					ImageFilter:     mockFilter,
					Mapper:          new(KeptnCloudEventMapper),
					K8s:             mockK8s,
					ErrorSender:     mockUniformErrorSender,
				}

				fakeKeptn := sdk.NewFakeKeptn("test-job-executor-service")
				fakeKeptn.AddTaskHandler("*", &sut)

				err := fakeKeptn.NewEvent(newEvent("../../test/events/action.triggered.json"))
				require.NoError(t, err)

				// TODO: weird that if connection to k8s fails we don't send back an error from the handling function
				for i, eventType := range test.expectedEventTypes {
					fakeKeptn.AssertSentEventType(t, i, eventType)
				}
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
				mockJobConfigReader.EXPECT().GetJobConfig("").Return(
					&config, "<config-hash-value>",
					nil,
				).Times(1)

				mockK8s := eventhandlerfake.NewMockK8s(mockCtrl)
				if len(test.expected.jobsInvocations) > 0 {
					mockK8s.EXPECT().ConnectToCluster().Return(nil).Times(1)
				} else {
					mockK8s.EXPECT().ConnectToCluster().Return(nil).Times(0)
				}

				mockUniformErrorSender := eventhandlerfake.NewMockErrorLogSender(mockCtrl)

				totalNoOfExpectedTasks := 0
				for _, invocation := range test.expected.jobsInvocations {
					for range invocation.taskNames {
						mockK8s.EXPECT().CreateK8sJob(
							gomock.Any(),
							gomock.Eq(k8sutils.JobDetails{
								Action:        &config.Actions[0],
								Task:          &config.Actions[0].Tasks[0],
								ActionIndex:   0,
								TaskIndex:     0,
								JobConfigHash: "<config-hash-value>",
							}),
							gomock.Any(),
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
					JobConfigReader: mockJobConfigReader,
					ServiceName:     "",
					JobSettings:     k8sutils.JobSettings{},
					ImageFilter:     mockFilter,
					Mapper:          new(KeptnCloudEventMapper),
					ErrorSender:     mockUniformErrorSender,
					K8s:             mockK8s,
				}

				fakeKeptn := sdk.NewFakeKeptn("test-job-executor-service")
				fakeKeptn.AddTaskHandler("*", &sut)

				err := fakeKeptn.NewEvent(newEvent(test.inputs.eventFile))
				require.NoError(t, err)

				for index, event := range test.expected.events {
					fakeKeptn.AssertSentEventType(t, index, event)
				}
			},
		)
	}
}

func TestStartK8s(t *testing.T) {
	k8sMock := createK8sMock(t)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFilter := eventhandlerfake.NewMockImageFilter(mockCtrl)
	mockJobConfigReader := eventhandlerfake.NewMockJobConfigReader(mockCtrl)
	mockUniformErrorSender := eventhandlerfake.NewMockErrorLogSender(mockCtrl)

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
				Namespace: "keptn2",
			},
		},
		Events: []config.Event{
			{
				Name: "sh.keptn.event.action.triggered",
			},
		},
	}

	mockJobConfigReader.EXPECT().GetJobConfig("").Return(
		&config.Config{
			Actions: []config.Action{action},
		}, "", nil,
	).Times(1)

	eh := EventHandler{
		JobConfigReader: mockJobConfigReader,
		ServiceName:     "test-jes",
		JobSettings:     k8sutils.JobSettings{},
		ImageFilter:     mockFilter,
		Mapper:          new(KeptnCloudEventMapper),
		K8s:             k8sMock,
		ErrorSender:     mockUniformErrorSender,
	}

	mockFilter.EXPECT().IsImageAllowed(gomock.Any()).Return(true).MinTimes(1)
	k8sMock.EXPECT().ConnectToCluster().Times(1)
	k8sMock.EXPECT().CreateK8sJob(
		gomock.Eq(jobName1), gomock.Eq(k8sutils.JobDetails{
			Action:        &action,
			Task:          &action.Tasks[0],
			ActionIndex:   0,
			TaskIndex:     0,
			JobConfigHash: "",
		}), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(),
	).Times(1)
	k8sMock.EXPECT().CreateK8sJob(
		gomock.Eq(jobName2), gomock.Eq(k8sutils.JobDetails{
			Action:        &action,
			Task:          &action.Tasks[1],
			ActionIndex:   0,
			TaskIndex:     1,
			JobConfigHash: "",
		}), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(),
	).Times(1)
	k8sMock.EXPECT().AwaitK8sJobDone(gomock.Eq(jobName1), 1006*time.Second, pollInterval,
		gomock.Any()).Times(1)
	k8sMock.EXPECT().AwaitK8sJobDone(gomock.Eq(jobName2), defaultMaxPollDuration, pollInterval, gomock.Any()).Times(1)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName1), gomock.Any()).Times(1)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName2), gomock.Any()).Times(1)

	fakeKeptn := sdk.NewFakeKeptn("test-job-executor-service")
	fakeKeptn.AddTaskHandler("*", &eh)

	err := fakeKeptn.NewEvent(newEvent("../../test/events/action.triggered.json"))
	require.NoError(t, err)

	fakeKeptn.AssertSentEventType(t, 0, "sh.keptn.event.action.started")
	fakeKeptn.AssertSentEventType(t, 1, "sh.keptn.event.action.finished")
}

func TestStartK8sJobSilent(t *testing.T) {
	k8sMock := createK8sMock(t)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockJobConfigReader := eventhandlerfake.NewMockJobConfigReader(mockCtrl)
	mockUniformErrorSender := eventhandlerfake.NewMockErrorLogSender(mockCtrl)

	eh := EventHandler{
		ServiceName:     "job-executor-service",
		ImageFilter:     acceptAllImagesFilter{},
		JobConfigReader: mockJobConfigReader,
		Mapper:          new(KeptnCloudEventMapper),
		K8s:             k8sMock,
		ErrorSender:     mockUniformErrorSender,
		JobSettings:     k8sutils.JobSettings{},
	}

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
		Events: []config.Event{
			{
				Name: "sh.keptn.event.action.triggered",
			},
		},
		Silent: true,
	}

	mockJobConfigReader.EXPECT().GetJobConfig("").Return(
		&config.Config{
			Actions: []config.Action{action},
		}, "", nil,
	).Times(1)

	k8sMock.EXPECT().ConnectToCluster().Times(1)
	k8sMock.EXPECT().CreateK8sJob(
		gomock.Eq(jobName1), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Times(1)
	k8sMock.EXPECT().CreateK8sJob(
		gomock.Eq(jobName2), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Times(1)
	k8sMock.EXPECT().AwaitK8sJobDone(gomock.Any(), defaultMaxPollDuration, pollInterval, "").Times(2)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName1), gomock.Any()).Times(1)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName2), gomock.Any()).Times(1)

	fakeKeptn := sdk.NewFakeKeptn("test-job-executor-service")
	fakeKeptn.AddTaskHandler("*", &eh)

	err := fakeKeptn.NewEvent(newEvent("../../test/events/action.triggered.json"))
	require.NoError(t, err)

	// Only one event means the finished event was skipped
	fakeKeptn.AssertNumberOfEventSent(t, 1)
}

func TestStartK8s_TestFinishedEvent(t *testing.T) {
	k8sMock := createK8sMock(t)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockJobConfigReader := eventhandlerfake.NewMockJobConfigReader(mockCtrl)
	mockUniformErrorSender := eventhandlerfake.NewMockErrorLogSender(mockCtrl)

	eh := EventHandler{
		ServiceName:     "job-executor-service",
		ImageFilter:     acceptAllImagesFilter{},
		JobConfigReader: mockJobConfigReader,
		Mapper:          new(KeptnCloudEventMapper),
		K8s:             k8sMock,
		ErrorSender:     mockUniformErrorSender,
		JobSettings:     k8sutils.JobSettings{},
	}

	action := config.Action{
		Name: "Run locust",
		Tasks: []config.Task{
			{
				Name: "Run locust healthy snack tests",
			},
		},
		Events: []config.Event{
			{
				Name: "sh.keptn.event.test.triggered",
			},
		},
	}

	mockJobConfigReader.EXPECT().GetJobConfig("").Return(
		&config.Config{
			Actions: []config.Action{action},
		}, "", nil,
	).Times(1)

	k8sMock.EXPECT().ConnectToCluster().Times(1)
	k8sMock.EXPECT().CreateK8sJob(
		gomock.Eq(jobName1), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Times(1)
	k8sMock.EXPECT().AwaitK8sJobDone(gomock.Eq(jobName1), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName1), gomock.Any()).Times(1)

	// set the global timezone for testing
	local, err := time.LoadLocation("UTC")
	require.NoError(t, err)
	time.Local = local

	fakeKeptn := sdk.NewFakeKeptn("test-job-executor-service")
	fakeKeptn.AddTaskHandler("*", &eh)

	err = fakeKeptn.NewEvent(newEvent("../../test/events/test.triggered.json"))
	require.NoError(t, err)

	fakeKeptn.AssertSentEventType(t, 0, keptnv2.GetStartedEventType(keptnv2.TestTaskName))
	fakeKeptn.AssertSentEventType(t, 1, keptnv2.GetFinishedEventType(keptnv2.TestTaskName))

	fakeKeptn.AssertSentEvent(t, 0, checkFinishedEvent)
	fakeKeptn.AssertSentEvent(t, 1, checkFinishedEvent)
}

func checkFinishedEvent(ce models.KeptnContextExtendedCE) bool {
	eventData := &keptnv2.TestFinishedEventData{}
	err := ce.DataAs(eventData)
	if err != nil {
		return false
	}

	return eventData.Status == keptnv2.StatusSucceeded
}

type disallowAllImagesFilter struct {
	ImageFilter
}

func (f disallowAllImagesFilter) IsImageAllowed(_ string) bool {
	return false
}

func TestExpectImageNotAllowedError(t *testing.T) {
	k8sMock := createK8sMock(t)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockJobConfigReader := eventhandlerfake.NewMockJobConfigReader(mockCtrl)
	mockUniformErrorSender := eventhandlerfake.NewMockErrorLogSender(mockCtrl)

	eh := EventHandler{
		ServiceName:     "job-executor-service",
		ImageFilter:     acceptAllImagesFilter{},
		JobConfigReader: mockJobConfigReader,
		Mapper:          new(KeptnCloudEventMapper),
		K8s:             k8sMock,
		ErrorSender:     mockUniformErrorSender,
		JobSettings:     k8sutils.JobSettings{},
	}

	notAllowedImageName := "alpine:latest"
	action := config.Action{
		Name: "Run some task with invalid image",
		Tasks: []config.Task{
			{
				Image: notAllowedImageName,
				Name:  "Run some image",
			},
		},
		Events: []config.Event{
			{
				Name: "sh.keptn.event.test.triggered",
			},
		},
	}

	mockJobConfigReader.EXPECT().GetJobConfig("").Return(
		&config.Config{
			Actions: []config.Action{action},
		}, "", nil,
	).Times(1)

	k8sMock.EXPECT().ConnectToCluster().Times(1)
	k8sMock.EXPECT().CreateK8sJob(
		gomock.Eq(jobName1), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Times(1)
	k8sMock.EXPECT().AwaitK8sJobDone(gomock.Eq(jobName1), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	k8sMock.EXPECT().GetLogsOfPod(gomock.Eq(jobName1), gomock.Any()).Times(1)

	// set the global timezone for testing
	local, err := time.LoadLocation("UTC")
	require.NoError(t, err)
	time.Local = local

	fakeKeptn := sdk.NewFakeKeptn("test-job-executor-service")
	fakeKeptn.AddTaskHandler("*", &eh)

	err = fakeKeptn.NewEvent(newEvent("../../test/events/test.triggered.json"))
	require.NoError(t, err)

	fakeKeptn.AssertSentEventType(t, 0, keptnv2.GetStartedEventType(keptnv2.TestTaskName))
	fakeKeptn.AssertSentEventType(t, 1, keptnv2.GetFinishedEventType(keptnv2.TestTaskName))

	fakeKeptn.AssertSentEvent(t, 0, checkFinishedEvent)
}

func checkFinishedEventImageNotAllowed(ce models.KeptnContextExtendedCE) bool {
	eventData := &keptnv2.TestFinishedEventData{}
	err := ce.DataAs(eventData)
	if err != nil {
		return false
	}

	if eventData.Status != keptnv2.StatusErrored {
		return false
	} else if eventData.Result != keptnv2.ResultFailed {
		return false
	}

	return true
}