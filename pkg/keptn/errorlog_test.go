package keptn

import (
	"encoding/json"
	"errors"
	"github.com/keptn/go-utils/pkg/sdk"
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/keptn/go-utils/pkg/api/models"
	"github.com/stretchr/testify/assert"
	"keptn-contrib/job-executor-service/pkg/keptn/fake"
)

func TestErrorWhenInitialCloudEventNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uniformClient := fake.NewMockUniformClient(ctrl)
	mockLogEventSender := fake.NewMockLogEventSender(ctrl)

	sut := NewErrorLogSender("foobar", uniformClient, mockLogEventSender)

	err := sut.SendErrorLogEvent(nil, errors.New("error text"))

	assert.ErrorIs(t, err, ErrorInitialCloudEventNotSpecified)

}

func TestErrorWhenProcessingErrorNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uniformClient := fake.NewMockUniformClient(ctrl)
	mockLogEventSender := fake.NewMockLogEventSender(ctrl)

	sut := NewErrorLogSender("foobar", uniformClient, mockLogEventSender)

	newEvent := sdk.KeptnEvent{}
	err := sut.SendErrorLogEvent(&newEvent, nil)

	assert.ErrorIs(t, err, ErrorProcessingErrorNotSpecified)

}

func TestErrorWhenGetRegistrationsFails(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uniformClient := fake.NewMockUniformClient(ctrl)
	getRegistrationError := errors.New("getRegistrations didn't work for some reason")
	uniformClient.EXPECT().GetRegistrations(gomock.Any(), gomock.Any()).Return(nil, getRegistrationError).Times(1)
	mockLogEventSender := fake.NewMockLogEventSender(ctrl)

	sut := NewErrorLogSender("", uniformClient, mockLogEventSender)

	newEvent := sdk.KeptnEvent{}
	err := sut.SendErrorLogEvent(&newEvent, errors.New("some error"))

	assert.Error(t, err)
	assert.ErrorIs(t, err, getRegistrationError)
}

func TestErrorWhenNoRegistrationIsReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uniformClient := fake.NewMockUniformClient(ctrl)
	uniformClient.EXPECT().GetRegistrations(gomock.Any(), gomock.Any()).Return([]*models.Integration{}, nil).Times(1)
	mockLogEventSender := fake.NewMockLogEventSender(ctrl)

	sut := NewErrorLogSender("foobar", uniformClient, mockLogEventSender)

	newEvent := sdk.KeptnEvent{}
	err := sut.SendErrorLogEvent(&newEvent, errors.New("some error"))

	assert.Error(t, err)
	assert.ErrorContains(t, err, "no registration found with name foobar")
}

func TestErrorWhenNoMatchingRegistrationIsReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uniformClient := fake.NewMockUniformClient(ctrl)
	uniformClient.EXPECT().GetRegistrations(gomock.Any(), gomock.Any()).Return(
		[]*models.Integration{
			{
				ID:   "idbazz",
				Name: "baz",
			},
		}, nil,
	).Times(1)

	mockLogEventSender := fake.NewMockLogEventSender(ctrl)

	sut := NewErrorLogSender("foobar", uniformClient, mockLogEventSender)

	newEvent := sdk.KeptnEvent{}
	err := sut.SendErrorLogEvent(&newEvent, errors.New("some error"))

	assert.Error(t, err)
	assert.ErrorContains(t, err, "no registration found with name foobar")
}

func TestErrorWhenMultipleMatchingRegistrationReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uniformClient := fake.NewMockUniformClient(ctrl)
	uniformClient.EXPECT().GetRegistrations(gomock.Any(), gomock.Any()).Return(
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

	keptnContext := "returnedEventContext"

	mockLogEventSender := fake.NewMockLogEventSender(ctrl)
	mockLogEventSender.EXPECT().Log(gomock.Any(), gomock.Any()).Times(2)
	mockLogEventSender.EXPECT().Flush(gomock.Any(), gomock.Any()).Times(2).Return(nil)

	sut := NewErrorLogSender("foobar", uniformClient, mockLogEventSender)

	eventType := "sh.keptn.test.triggered"
	newEvent := sdk.KeptnEvent{
		Shkeptncontext: keptnContext,
		Type:           &eventType,
	}

	err := sut.SendErrorLogEvent(&newEvent, errors.New("some error"))
	assert.NoError(t, err)
}

func TestSendErrorLogHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uniformClient := fake.NewMockUniformClient(ctrl)
	uniformClient.EXPECT().GetRegistrations(gomock.Any(), gomock.Any()).Return(
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

	initialCloudEvent := sdk.KeptnEvent{}
	bytes, err := os.ReadFile("../../test/events/action.triggered.json")
	require.NoError(t, err)
	json.Unmarshal(bytes, &initialCloudEvent)

	testError := errors.New("some job executor error")

	// ErrorCloudEventData cannot be compared since models.LogEntry{} has no matcher
	mockLogEventSender := fake.NewMockLogEventSender(ctrl)
	mockLogEventSender.EXPECT().Log(gomock.AssignableToTypeOf([]models.LogEntry{}), gomock.Any()).Times(1)
	mockLogEventSender.EXPECT().Flush(gomock.Any(), gomock.Any()).Times(1).Return(nil)

	sut := NewErrorLogSender("bar", uniformClient, mockLogEventSender)

	err = sut.SendErrorLogEvent(&initialCloudEvent, testError)

	require.NoError(t, err)
}
