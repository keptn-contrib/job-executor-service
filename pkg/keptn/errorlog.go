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

var /*const*/ ErrorInitialCloudEventNotSpecified = errors.New("initial cloudevent not specified")
var /*const*/ ErrorProcessingErrorNotSpecified = errors.New("processing error is nil")

type UniformClient interface {
	GetRegistrations() ([]*models.Integration, error)
}

type CloudEventSender interface {
	SendCloudEvent(event cloudevents.Event) error
}

//go:generate mockgen -destination=fake/errorlog_mock.go -package=fake .  UniformClient,CloudEventSender

type ErrorLogSender struct {
	uniformHandler  UniformClient
	ceSender        CloudEventSender
	integrationName string
}

func NewErrorLogSender(integrationName string, uniformClient UniformClient, sender CloudEventSender) *ErrorLogSender {
	return &ErrorLogSender{
		uniformHandler:  uniformClient,
		ceSender:        sender,
		integrationName: integrationName,
	}
}

type ErrorData struct {
	Message       string `json:"message"`
	IntegrationID string `json:"integrationid"`
	Task          string `json:"task,omitempty"`
}

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

	var integrationId string
	for _, registration := range registrations {
		if registration.Name == els.integrationName {
			if integrationId != "" {
				return fmt.Errorf("found multiple uniform registrations with name %s", els.integrationName)
			}
			integrationId = registration.ID
		}
	}

	if integrationId == "" {
		return fmt.Errorf("no registration found with name %s", els.integrationName)
	}

	errorCloudEvent, err := createErrorLogCloudEvent(integrationId, initialCloudEvent, applicationError)
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
	ev.SetDataContentType(cloudevents.ApplicationJSON)
	ev.SetType(errorType)

	keptnCtx, err := types.ToString(initialEvent.Extensions()["shkeptncontext"])
	if err != nil {
		return ev, fmt.Errorf("unable to extract keptnshcontext from initial cloud event: %w", err)
	}

	ev.SetExtension("shkeptncontext", keptnCtx)

	err = ev.SetData(cloudevents.ApplicationJSON, errorData)
	if err != nil {
		return ev, fmt.Errorf("could not marshal cloud event payload: %v", err)
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
