package config

import (
	"fmt"
	"log"
)

const jobConfigResourceName = "job/config.yaml"

const minTTLSecondsAfterFinished = int32(60)

//go:generate mockgen -destination=fake/reader_mock.go -package=fake . KeptnResourceService

// KeptnResourceService defines the contract used by JobConfigReader to retrieve a resource from keptn (using project,
// service, stage from context)
type KeptnResourceService interface {
	GetKeptnResource(resource string) ([]byte, error)
}

// JobConfigReader retrieves and parses job configuration from Keptn
type JobConfigReader struct {
	Keptn KeptnResourceService
}

// GetJobConfig retrieves job/config.yaml resource from keptn and parses it into a Config struct.
// In case of error retrieving the resource or parsing the yaml it will return (nil,
// error) with the original error correctly wrapped in the local one
func (jcr *JobConfigReader) GetJobConfig() (*Config, error) {
	resource, err := jcr.Keptn.GetKeptnResource(jobConfigResourceName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving job config: %w", err)
	}

	configuration, err := NewConfig(resource)
	if err != nil {
		log.Printf("Could not parse config: %s", err)
		log.Printf("The config was: %s", string(resource))
		return nil, fmt.Errorf("error parsing job configuration: %w", err)
	}

	// Validate / Correct the Job cleanup TTL to 60s to avoid a job cleanup before the
	// job-executor-service can collect the logs of a completed job
	for _, action := range configuration.Actions {
		for taskIndex := range action.Tasks {
			task := &action.Tasks[taskIndex]

			if task.TTLSecondsAfterFinished != nil && *task.TTLSecondsAfterFinished < minTTLSecondsAfterFinished {
				log.Printf("Warning: Correcting TTLSecondsAfterFinished in action '%s' for task '%s' to %d!",
					action.Name, task.Name, minTTLSecondsAfterFinished,
				)

				var minJobTTLSeconds = minTTLSecondsAfterFinished
				task.TTLSecondsAfterFinished = &minJobTTLSeconds
			}
		}
	}

	return configuration, nil
}
