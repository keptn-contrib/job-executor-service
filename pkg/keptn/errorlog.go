package keptn

import (
	"errors"
	"fmt"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/keptn/go-utils/pkg/api/models"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

const errorType = "sh.keptn.log.error"

// ErrorInitialCloudEventNotSpecified is returned when the cloud event passed to ErrorLogSender#SendErrorLogEvent is nil
var /*const*/ ErrorInitialCloudEventNotSpecified = errors.New("initial cloudevent not specified")

// ErrorProcessingErrorNotSpecified is returned when the error passed to ErrorLogSender#SendErrorLogEvent is nil
var /*const*/ ErrorProcessingErrorNotSpecified = errors.New("processing error is nil")

// UniformClient represents the interface implemented  by the Keptn Uniform API client
type UniformClient interface {
	GetRegistrations() ([]*models.Integration, error)
}

// CloudEventSender represents the interface implemented by the Keptn API client for sending cloudevents
type CloudEventSender interface {
	SendCloudEvent(event cloudevents.Event) error
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
func (els *ErrorLogSender) SendErrorLogEvent(initialCloudEvent *cloudevents.Event, applicationError error) error {

	if initialCloudEvent == nil {
		return ErrorInitialCloudEventNotSpecified
	}

	if applicationError == nil {
		return ErrorProcessingErrorNotSpecified
	}

	registrations, err := els.uniformHandler.GetRegistrations()
	if err != nil {
		return fmt.Errorf("error retrieving uniform registrations: %w", err)
	}

	var integrationID string
	for _, registration := range registrations {
		if registration.Name == els.integrationName {
			if integrationID != "" {
				return fmt.Errorf("found multiple uniform registrations with name %s", els.integrationName)
			}
			integrationID = registration.ID
		}
	}

	if integrationID == "" {
		return fmt.Errorf("no registration found with name %s", els.integrationName)
	}

	errorCloudEvent, err := createErrorLogCloudEvent(integrationID, initialCloudEvent, applicationError)
	if err != nil {
		return fmt.Errorf("unable to create error log cloudevent: %w", err)
	}

	els.ceSender.SendCloudEvent(errorCloudEvent)

	return nil
}

func createErrorLogCloudEvent(
	integrationID string, initialEvent *cloudevents.Event, err error,
) (cloudevents.Event, error) {
	errorData := ErrorData{
		Message:       err.Error(),
		IntegrationID: integrationID,
		Task:          getTaskFromEvent(initialEvent.Type()),
	}

	ev := cloudevents.NewEvent()
	ev.SetSource(initialEvent.Source())
	ev.SetExtension("triggeredid", initialEvent.ID())
	ev.SetDataContentType(cloudevents.ApplicationJSON)
	ev.SetType(errorType)

	keptnCtx, err := types.ToString(initialEvent.Extensions()["shkeptncontext"])
	if err != nil {
		return ev, fmt.Errorf("unable to extract keptnshcontext from initial cloud event: %w", err)
	}

	ev.SetExtension("shkeptncontext", keptnCtx)

	err = ev.SetData(cloudevents.ApplicationJSON, errorData)
	if err != nil {
		return ev, fmt.Errorf("could not marshal cloud event payload: %w", err)
	}

	return ev, nil
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
