package eventhandler

import (
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeEventPayloadAsInterface(t *testing.T) {

	context := spec.V1.NewContext()
	context.SetID("0123")
	context.SetSource("sourcysource")
	now := time.Now()
	context.SetTime(now)
	context.SetExtension("shkeptncontext", interface{}("mycontext"))

	event := event.Event{
		Context:     context,
		DataEncoded: []byte(testEvent),
	}

	mapper := new(KeptnCloudEventMapper)
	eventPayloadAsInterface, err := mapper.Map(event)
	require.NoError(t, err)

	assert.Equal(t, eventPayloadAsInterface["id"], "0123")
	assert.Equal(t, eventPayloadAsInterface["source"], "sourcysource")
	assert.Equal(t, eventPayloadAsInterface["time"], now)
	assert.Equal(t, eventPayloadAsInterface["shkeptncontext"], "mycontext")

	data := eventPayloadAsInterface["data"]
	dataAsMap := data.(map[string]interface{})

	assert.Equal(t, dataAsMap["project"], "sockshop")
}
