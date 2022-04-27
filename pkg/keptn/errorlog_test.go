package keptn

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/keptn/go-utils/pkg/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"keptn-contrib/job-executor-service/pkg/keptn/fake"
)

func TestErrorWhenInitialCloudEventNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uniformClient := fake.NewMockUniformClient(ctrl)
	mockCloudEventSender := fake.NewMockCloudEventSender(ctrl)

	sut := NewErrorLogSender("foobar", uniformClient, mockCloudEventSender)

	err := sut.SendErrorLogEvent(nil, errors.New("error text"))

	assert.ErrorIs(t, err, ErrorInitialCloudEventNotSpecified)

}

func TestErrorWhenProcessingErrorNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uniformClient := fake.NewMockUniformClient(ctrl)
	mockCloudEventSender := fake.NewMockCloudEventSender(ctrl)

	sut := NewErrorLogSender("foobar", uniformClient, mockCloudEventSender)

	newEvent := cloudevents.NewEvent()
	err := sut.SendErrorLogEvent(&newEvent, nil)

	assert.ErrorIs(t, err, ErrorProcessingErrorNotSpecified)

}

func TestErrorWhenGetRegistrationsFails(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uniformClient := fake.NewMockUniformClient(ctrl)
	getRegistrationError := errors.New("getRegistrations didn't work for some reason")
	uniformClient.EXPECT().GetRegistrations().Return(nil, getRegistrationError).Times(1)
	mockCloudEventSender := fake.NewMockCloudEventSender(ctrl)

	sut := NewErrorLogSender("", uniformClient, mockCloudEventSender)

	newEvent := cloudevents.NewEvent()
	err := sut.SendErrorLogEvent(&newEvent, errors.New("some error"))

	assert.Error(t, err)
	assert.ErrorIs(t, err, getRegistrationError)
}

func TestErrorWhenNoRegistrationIsReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uniformClient := fake.NewMockUniformClient(ctrl)
	uniformClient.EXPECT().GetRegistrations().Return([]*models.Integration{}, nil).Times(1)
	mockCloudEventSender := fake.NewMockCloudEventSender(ctrl)
	sut := NewErrorLogSender("foobar", uniformClient, mockCloudEventSender)

	newEvent := cloudevents.NewEvent()
	err := sut.SendErrorLogEvent(&newEvent, errors.New("some error"))

	assert.Error(t, err)
	assert.ErrorContains(t, err, "no registration found with name foobar")
}

func TestErrorWhenNoMatchingRegistrationIsReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uniformClient := fake.NewMockUniformClient(ctrl)
	uniformClient.EXPECT().GetRegistrations().Return(
		[]*models.Integration{
			{
				ID:   "idbazz",
				Name: "baz",
			},
		}, nil,
	).Times(1)

	mockCloudEventSender := fake.NewMockCloudEventSender(ctrl)
	sut := NewErrorLogSender("foobar", uniformClient, mockCloudEventSender)

	newEvent := cloudevents.NewEvent()
	err := sut.SendErrorLogEvent(&newEvent, errors.New("some error"))

	assert.Error(t, err)
	assert.ErrorContains(t, err, "no registration found with name foobar")
}

func TestErrorWhenMultipleMatchingRegistrationReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uniformClient := fake.NewMockUniformClient(ctrl)
	uniformClient.EXPECT().GetRegistrations().Return(
		[]*models.Integration{
			{
				ID:   "idfoobar1",
				Name: "foobar",
			},
			{
				ID:   "idbazz",
				Name: "baz",
			},
			{
				ID:   "idfoobar2",
				Name: "foobar",
			},
		}, nil,
	).Times(1)

	mockCloudEventSender := fake.NewMockCloudEventSender(ctrl)
	sut := NewErrorLogSender("foobar", uniformClient, mockCloudEventSender)

	newEvent := cloudevents.NewEvent()
	err := sut.SendErrorLogEvent(&newEvent, errors.New("some error"))

	assert.Error(t, err)
	assert.ErrorContains(t, err, "found multiple uniform registrations with name foobar")
}

func TestSendErrorLogHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uniformClient := fake.NewMockUniformClient(ctrl)
	uniformClient.EXPECT().GetRegistrations().Return(
		[]*models.Integration{
			{
				ID:   "idfoo",
				Name: "foo",
			},
			{
				ID:   "idbar",
				Name: "bar",
			},
			{
				ID:   "idbaz",
				Name: "baz",
			},
		},
		nil,
	).Times(1)

	initialCloudEvent := cloudevents.NewEvent()
	bytes, err := os.ReadFile("../../test/events/action.triggered.json")
	require.NoError(t, err)
	json.Unmarshal(bytes, &initialCloudEvent)

	expectedErrorCloudEvent := cloudevents.NewEvent()
	expectedErrorCloudEvent.SetType(errorType)
	expectedErrorCloudEvent.SetSource("test-events")
	expectedKeptnCtx, err := types.ToString(
		initialCloudEvent.Extensions()["shkeptncontext"],
	)
	require.NoError(t, err)

	expectedErrorCloudEvent.SetExtension("shkeptncontext", expectedKeptnCtx)

	testError := errors.New("some job executor error")
	expectedErrorData := ErrorData{
		Message:       testError.Error(),
		IntegrationID: "idbar",
		Task:          "action",
	}

	expectedErrorCloudEvent.SetData(cloudevents.ApplicationJSON, expectedErrorData)

	mockCloudEventSender := fake.NewMockCloudEventSender(ctrl)
	mockCloudEventSender.EXPECT().SendCloudEvent(gomock.Eq(expectedErrorCloudEvent)).Times(1)

	sut := NewErrorLogSender("bar", uniformClient, mockCloudEventSender)

	err = sut.SendErrorLogEvent(&initialCloudEvent, testError)

	require.NoError(t, err)
}
