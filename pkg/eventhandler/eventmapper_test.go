package eventhandler

import (
	"encoding/json"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/keptn/go-utils/pkg/sdk"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeEventPayloadAsInterface(t *testing.T) {
	source := "sourcysource"
	now := time.Now()

	eventData := &keptnv2.EventData{}
	err := json.Unmarshal([]byte(testEvent), eventData)
	require.NoError(t, err)

	keptnEvent := sdk.KeptnEvent{
		ID:             "0123",
		Source:         &source,
		Time:           now,
		Shkeptncontext: "mycontext",
		Data:           eventData,
	}

	mapper := new(KeptnCloudEventMapper)
	eventPayloadAsInterface, err := mapper.Map(keptnEvent)
	require.NoError(t, err)

	assert.Equal(t, "0123", eventPayloadAsInterface["id"])
	assert.Equal(t, now, eventPayloadAsInterface["time"])
	assert.Equal(t, "mycontext", eventPayloadAsInterface["shkeptncontext"])

	data := eventPayloadAsInterface["data"]
	dataAsMap := data.(interface{}).(map[string]interface{})

	assert.Equal(t, dataAsMap["project"], "sockshop")
}
