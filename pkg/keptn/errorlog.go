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
	"time"
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

// LogEventSender represents the interface implemented by the Keptn API client for sending log messages
type LogEventSender interface {
	// Log appends the specified logs to the log cache.
	Log(logs []models.LogEntry, opts api.LogsLogOptions)

	// Flush flushes the log cache.
	Flush(ctx context.Context, opts api.LogsFlushOptions) error
}

//go:generate mockgen -destination=fake/errorlog_mock.go -package=fake .  UniformClient,LogEventSender

// ErrorLogSender creates and sends the error log cloudevents for the registered job-executor-service extension
type ErrorLogSender struct {
	uniformHandler  UniformClient
	logSender       LogEventSender
	integrationName string
}

// NewErrorLogSender returns an initialized ErrorLogSender
func NewErrorLogSender(integrationName string, uniformClient UniformClient, sender LogEventSender) *ErrorLogSender {
	return &ErrorLogSender{
		uniformHandler:  uniformClient,
		logSender:       sender,
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
			errorLog := createErrorLog(registration.ID, initialCloudEvent, applicationError)

			els.logSender.Log([]models.LogEntry{errorLog}, api.LogsLogOptions{})
			eventErr := els.logSender.Flush(context.Background(), api.LogsFlushOptions{})
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

func createErrorLog(
	integrationID string, initialEvent *sdk.KeptnEvent, err error,
) models.LogEntry {
	logEntry := models.LogEntry{
		GitCommitID:   initialEvent.GitCommitID,
		KeptnContext:  initialEvent.Shkeptncontext,
		Message:       err.Error(),
		Time:          time.Now(),
		Task:          getTaskFromEvent(*initialEvent.Type),
		IntegrationID: integrationID,
		TriggeredID:   initialEvent.Triggeredid,
	}

	return logEntry
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
