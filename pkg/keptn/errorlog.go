package keptn

import (
	"context"
	"errors"
	"fmt"
	"github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils/v2"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/keptn/go-utils/pkg/sdk"
	"log"
)

const errorType = "sh.keptn.log.error"

// ErrorInitialCloudEventNotSpecified is returned when the cloud event passed to ErrorLogSender#SendErrorLogEvent is nil
var /*const*/ ErrorInitialCloudEventNotSpecified = errors.New("initial cloudevent not specified")

// ErrorProcessingErrorNotSpecified is returned when the error passed to ErrorLogSender#SendErrorLogEvent is nil
var /*const*/ ErrorProcessingErrorNotSpecified = errors.New("processing error is nil")

// UniformClient represents the interface implemented  by the Keptn Uniform API client
type UniformClient interface {
	GetRegistrations(ctx context.Context, opts api.UniformGetRegistrationsOptions) ([]*models.Integration, error)
}

// CloudEventSender represents the interface implemented by the Keptn API client for sending cloudevents
type CloudEventSender interface {
	SendEvent(ctx context.Context, event models.KeptnContextExtendedCE, opts api.APISendEventOptions) (*models.EventContext, *models.Error)
}

//go:generate mockgen -destination=fake/errorlog_mock.go -package=fake .  UniformClient,CloudEventSender

// ErrorLogSender creates and sends the error log cloudevents for the registered job-executor-service extension
type ErrorLogSender struct {
	uniformHandler  UniformClient
	ceSender        CloudEventSender
	integrationName string
}

// NewErrorLogSender returns an initialized ErrorLogSender
func NewErrorLogSender(integrationName string, uniformClient UniformClient, sender CloudEventSender) *ErrorLogSender {
	return &ErrorLogSender{
		uniformHandler:  uniformClient,
		ceSender:        sender,
		integrationName: integrationName,
	}
}

// ErrorData represents the cloudevent payload of the error log. See https://github.com/keptn/spec/blob/master/cloudevents.md#error-log
type ErrorData struct {
	Message       string `json:"message"`
	IntegrationID string `json:"integrationid"`
	Task          string `json:"task,omitempty"`
}

// SendErrorLogEvent will retrieve the current registration for the job-executor-service,
// create a cloudevent of type sh.keptn.log.error and send it back to keptn using the information retrieved from the
// triggering cloud event and the encountered error.
// If retrieving the integration registration or sending the error log fails, an error will be returned
func (els *ErrorLogSender) SendErrorLogEvent(initialCloudEvent *sdk.KeptnEvent, applicationError error) error {

	if initialCloudEvent == nil {
		return ErrorInitialCloudEventNotSpecified
	}

	if applicationError == nil {
		return ErrorProcessingErrorNotSpecified
	}

	registrations, err := els.uniformHandler.GetRegistrations(context.Background(), api.UniformGetRegistrationsOptions{})
	if err != nil {
		return fmt.Errorf("error retrieving uniform registrations: %w", err)
	}

	sendEvent := false
	for _, registration := range registrations {
		if registration.Name == els.integrationName {
			errorCloudEvent, err := createErrorLogCloudEvent(registration.ID, initialCloudEvent, applicationError)
			if err != nil {
				log.Printf("unable to create error log cloudevent %+v: %+v", initialCloudEvent, err)
				continue
			}

			_, eventErr := els.ceSender.SendEvent(context.Background(), errorCloudEvent, api.APISendEventOptions{})
			if eventErr == nil {
				sendEvent = true
			}

		}
	}

	if sendEvent {
		return nil
	}

	return fmt.Errorf("no registration found with name %s", els.integrationName)
}

func createErrorLogCloudEvent(
	integrationID string, initialEvent *sdk.KeptnEvent, err error,
) (models.KeptnContextExtendedCE, error) {
	errorData := ErrorData{
		Message:       err.Error(),
		IntegrationID: integrationID,
		Task:          getTaskFromEvent(*initialEvent.Type),
	}

	eventType := errorType
	event := models.KeptnContextExtendedCE{
		Data:           errorData,
		Source:         initialEvent.Source,
		Type:           &eventType,
		Triggeredid:    initialEvent.ID,
		Shkeptncontext: initialEvent.Shkeptncontext,
	}

	return event, nil
}

func getTaskFromEvent(eventType string) string {
	if !keptnv2.IsTaskEventType(eventType) {
		return eventType
	}

	taskName, _, err := keptnv2.ParseTaskEventType(eventType)
	if err != nil {
		log.Printf("could not extract task name from event type: %s, will set it to full type", eventType)
		return eventType
	}

	return taskName
}
