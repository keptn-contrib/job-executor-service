package config

import (
	"fmt"
	"golang.org/x/crypto/sha3"
	"log"
)

const jobConfigResourceName = "job/config.yaml"

//go:generate mockgen -destination=fake/reader_mock.go -package=fake . KeptnResourceService

// KeptnResourceService defines the contract used by JobConfigReader to retrieve a resource from Keptn
// The project, service, stage environment variables are taken from the context of the ResourceService (Event)
type KeptnResourceService interface {
	// GetServiceResource returns the service level resource
	GetServiceResource(resource string, gitCommitID string) ([]byte, error)

	// GetProjectResource returns the resource that was defined on project level
	GetProjectResource(resource string, gitCommitID string) ([]byte, error)

	// GetStageResource returns the resource that was defined in the stage
	GetStageResource(resource string, gitCommitID string) ([]byte, error)
}

// JobConfigReader retrieves and parses job configuration from Keptn
type JobConfigReader struct {
	Keptn KeptnResourceService
}

// FindJobConfigResource searches for the job configuration resource in the service, stage and then the project
// and returns the content of the first resource that is found
func (jcr *JobConfigReader) FindJobConfigResource(gitCommitID string) ([]byte, error) {
	if config, err := jcr.Keptn.GetServiceResource(jobConfigResourceName, gitCommitID); err == nil {
		return config, nil
	}

	if config, err := jcr.Keptn.GetStageResource(jobConfigResourceName, gitCommitID); err == nil {
		return config, nil
	}

	// FIXME: Since the resource service uses different branches, the commitID may not be in the main
	//        branch and therefore it's not possible to query the project fallback configuration!
	if config, err := jcr.Keptn.GetProjectResource(jobConfigResourceName, ""); err == nil {
		return config, nil
	}

	// TODO: Improve error handling:
	return nil, fmt.Errorf("unable to find job configuration")
}

// GetJobConfig retrieves job/config.yaml resource from keptn and parses it into a Config struct.
// Additionally, also the SHA1 hash of the retrieved configuration will be returned.
// In case of error retrieving the resource or parsing the yaml it will return (nil,
// error) with the original error correctly wrapped in the local one
func (jcr *JobConfigReader) GetJobConfig(gitCommitID string) (*Config, string, error) {

	resource, err := jcr.FindJobConfigResource(gitCommitID)
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

	return configuration, resourceHash, nil
}
