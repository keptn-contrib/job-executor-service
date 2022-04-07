package eventhandler

import (
	"encoding/json"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// KeptnCloudEventMapper is a simple mapper from cloudevent to map[string]interface{}
// used to parse generic JSON from the cloudevent data.
type KeptnCloudEventMapper struct {
	EventMapper
}

// Map transforms a cloud event into a generic map[string]interface{} as defined by EventMapper.Map
func (kcem *KeptnCloudEventMapper) Map(ce cloudevents.Event) (map[string]interface{}, error) {
	// we do the same in main#processKeptnCloudEvent:L82-85 but we use a &keptnv2.EventData{} there...
	var eventDataAsInterface interface{}
	err := json.Unmarshal(ce.Data(), &eventDataAsInterface)
	if err != nil {
		log.Printf("failed to convert incoming cloudevent: %v", err)
		return nil, err
	}

	extension, _ := ce.Context.GetExtension("shkeptncontext")
	shKeptnContext := extension.(string)

	eventAsInterface := make(map[string]interface{})
	eventAsInterface["id"] = ce.ID()
	eventAsInterface["shkeptncontext"] = shKeptnContext
	eventAsInterface["time"] = ce.Time()
	eventAsInterface["source"] = ce.Source()
	eventAsInterface["data"] = eventDataAsInterface
	eventAsInterface["specversion"] = ce.SpecVersion()
	eventAsInterface["type"] = ce.Type()

	return eventAsInterface, nil
}
