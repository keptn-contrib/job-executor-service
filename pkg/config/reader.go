package config

import (
	"fmt"
	"log"

	"golang.org/x/crypto/sha3"

	"keptn-contrib/job-executor-service/pkg/config/signing"
)

const jobConfigResourceName = "job/config.yaml"
const jobConfigSignatureResourceName = "job/config.yaml.sig"
const jobConfigAllowedSignersResourceName = "job/config.yaml.allowed_signers"

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

	jobConfigSignatureBytes, err := jcr.Keptn.GetKeptnResource(jobConfigSignatureResourceName)
	if err != nil {
		return nil, "", fmt.Errorf("error retrieving job config signature: %w", err)
	}

	// TODO this probably should not be stored together with the signature (
	// if git repo is compromised an attacker can easily attach another key here)
	// it should probably be stored within the k8s cluster (configMap probably)

	jobConfigAllowedSignersBytes, err := jcr.Keptn.GetKeptnResource(jobConfigAllowedSignersResourceName)
	if err != nil {
		return nil, "", fmt.Errorf("error retrieving job config allowed signers: %w", err)
	}

	err = signing.VerifyJobConfig(jobConfigBytes, jobConfigSignatureBytes, jobConfigAllowedSignersBytes)
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
