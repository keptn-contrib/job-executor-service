package config

import (
	"fmt"
	"golang.org/x/crypto/sha3"
	"log"
)

const jobConfigResourceName = "job/config.yaml"

//go:generate mockgen -destination=fake/reader_mock.go -package=fake . KeptnResourceService

// KeptnResourceService defines the contract used by JobConfigReader to retrieve a resource from keptn (using project,
// service, stage from context)
type KeptnResourceService interface {
	GetResource(resource string, gitCommitId string) ([]byte, error)
}

// JobConfigReader retrieves and parses job configuration from Keptn
type JobConfigReader struct {
	Keptn KeptnResourceService
}

// GetJobConfig retrieves job/config.yaml resource from keptn and parses it into a Config struct.
// Additionally, also the SHA1 hash of the retrieved configuration will be returned.
// In case of error retrieving the resource or parsing the yaml it will return (nil,
// error) with the original error correctly wrapped in the local one
func (jcr *JobConfigReader) GetJobConfig(gitCommitId string) (*Config, string, error) {

	resource, err := jcr.Keptn.GetResource(jobConfigResourceName, gitCommitId)
	if err != nil {
		return nil, "", fmt.Errorf("error retrieving job config: %w", err)
	}

	hasher := sha3.New224()
	hasher.Write(resource)
	resourceHashBytes := hasher.Sum(nil)
	resourceHash := fmt.Sprintf("%x", resourceHashBytes)

	configuration, err := NewConfig(resource)
	if err != nil {
		log.Printf("Could not parse config: %s", err)
		log.Printf("The config was: %s", string(resource))
		return nil, "", fmt.Errorf("error parsing job configuration: %w", err)
	}

	log.Printf("config: SHA3: %s\nGIT :%s\n%s\n", resourceHash, gitCommitId, resource)

	return configuration, resourceHash, nil
}
