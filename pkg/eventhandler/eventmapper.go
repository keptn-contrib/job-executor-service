package eventhandler

import (
	"encoding/json"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type KeptnCloudEventMapper struct {
	EventMapper
}

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
