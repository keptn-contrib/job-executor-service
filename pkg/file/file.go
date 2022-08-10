package file

import (
	"fmt"
	"keptn-contrib/job-executor-service/pkg/config"
	"log"
	"path/filepath"

	"github.com/spf13/afero"
)

// MountFiles requests all specified files of a task from the keptn configuration service and copies them to /keptn
func MountFiles(actionName string, taskName string, fs afero.Fs, jcr config.JobConfigReader) error {
	configuration, _, err := jcr.GetJobConfig("")
	if err != nil {
		return fmt.Errorf("could not find config for job-executor-service: %v", err)
	}

	found, action := configuration.FindActionByName(actionName)
	if !found {
		return fmt.Errorf("no action found with name '%s'", actionName)
	}

	found, task := action.FindTaskByName(taskName)
	if !found {
		return fmt.Errorf("no task found with name '%s'", taskName)
	}

	for _, resourcePath := range task.Files {
		allServiceResources, err := jcr.Keptn.GetAllKeptnResources(resourcePath)
		if err != nil {
			return fmt.Errorf("could not retrieve resources for task '%v': %v", taskName, err)
		}

		// If the given resource is a folder, all files contained in the folder have to be copied over
		// to the filesystem of the workload
		for resourceURI, resourceContent := range allServiceResources {
			// Our mount starts with /keptn
			dir := filepath.Join("/keptn", filepath.Dir(resourceURI))
			fullFilePath := filepath.Join("/keptn", resourceURI)

			err := fs.MkdirAll(dir, 0700)
			if err != nil {
				return fmt.Errorf("could not create directory %s for file %s: %v", dir, resourceURI, err)
			}

			file, err := fs.Create(fullFilePath)
			if err != nil {
				return fmt.Errorf("could not create file %s: %v", resourceURI, err)
			}

			_, err = file.Write(resourceContent)
			defer func() {
				err = file.Close()
				if err != nil {
					log.Printf("could not close file %s: %v", file.Name(), err)
				}
			}()

			if err != nil {
				return fmt.Errorf("could not write to file %s: %v", fullFilePath, err)
			}

			log.Printf("successfully moved file %s to %s", resourceURI, fullFilePath)
		}

		if len(allServiceResources) == 0 {
			return fmt.Errorf("could not find file or directory %s for task %s", resourcePath, taskName)
		}
	}

	return nil
}
