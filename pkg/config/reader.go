package config

import (
	"fmt"
	"log"

	"golang.org/x/crypto/sha3"
)

const jobConfigResourceName = "job/config.yaml"

//go:generate mockgen -destination=fake/reader_mock.go -package=fake . KeptnResourceService,JobConfigVerifier

// KeptnResourceService defines the contract used by JobConfigReader to retrieve a resource from keptn (using project,
// service, stage from context)
type KeptnResourceService interface {
	GetKeptnResource(resource string) ([]byte, error)
}

type JobConfigVerifier interface {
	VerifyJobConfig(jobConfigBytes []byte) error
}

// JobConfigReader retrieves and parses job configuration from Keptn
type JobConfigReader struct {
	Keptn    KeptnResourceService
	Verifier JobConfigVerifier
}

// GetJobConfig retrieves job/config.yaml resource from keptn and parses it into a Config struct.
// Additionally, also the SHA1 hash of the retrieved configuration will be returned.
// In case of error retrieving the resource or parsing the yaml it will return (nil,
// error) with the original error correctly wrapped in the local one
func (jcr *JobConfigReader) GetJobConfig() (*Config, string, error) {
	jobConfigBytes, err := jcr.Keptn.GetKeptnResource(jobConfigResourceName)
	if err != nil {
		return nil, "", fmt.Errorf("error retrieving job config: %w", err)
	}

	hasher := sha3.New224()
	hasher.Write(jobConfigBytes)
	resourceHashBytes := hasher.Sum(nil)
	resourceHash := fmt.Sprintf("%x", resourceHashBytes)

	err = jcr.Verifier.VerifyJobConfig(jobConfigBytes)
	if err != nil {
		return nil, "", fmt.Errorf("error validating job config signature: %w", err)
	}

	configuration, err := NewConfig(jobConfigBytes)
	if err != nil {
		log.Printf("Could not parse config: %s", err)
		log.Printf("The config was: %s", string(jobConfigBytes))
		return nil, "", fmt.Errorf("error parsing job configuration: %w", err)
	}
	return configuration, resourceHash, nil
}
