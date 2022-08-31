package eventhandler

import (
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/keptn/go-utils/pkg/sdk"
	"log"
)

// KeptnCloudEventMapper is a simple mapper from cloudevent to map[string]interface{}
// used to parse generic JSON from the cloudevent data.
type KeptnCloudEventMapper struct {
	EventMapper
}

// Map transforms a cloud event into a generic map[string]interface{} as defined by EventMapper.Map
func (kcem *KeptnCloudEventMapper) Map(ce sdk.KeptnEvent) (map[string]interface{}, error) {
	// we do the same in main#processKeptnCloudEvent:L82-85 but we use a &keptnv2.EventData{} there...
	var eventDataAsInterface interface{}
	err := keptnv2.Decode(ce.Data, &eventDataAsInterface)
	if err != nil {
		log.Printf("failed to convert incoming cloudevent: %v", err)
		return nil, err
	}

	shKeptnContext := ce.Shkeptncontext

	eventAsInterface := make(map[string]interface{})
	eventAsInterface["id"] = ce.ID
	eventAsInterface["shkeptncontext"] = shKeptnContext
	eventAsInterface["time"] = ce.Time
	eventAsInterface["source"] = ce.Source
	eventAsInterface["data"] = eventDataAsInterface
	eventAsInterface["specversion"] = ce.Specversion
	eventAsInterface["type"] = ce.Type
	eventAsInterface["gitcommitid"] = ce.GitCommitID

	return eventAsInterface, nil
}
