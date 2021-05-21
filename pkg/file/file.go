package file

import (
	"didiladi/job-executor-service/pkg/config"
	"didiladi/job-executor-service/pkg/keptn"
	"fmt"
	"log"
	"net/url"
	"path/filepath"

	"github.com/spf13/afero"
)

// MountFiles requests all specified files of a task from the keptn configuration service and copies them to /keptn
func MountFiles(actionName string, taskName string, fs afero.Fs, configService keptn.ConfigService) error {

	// https://github.com/keptn/keptn/issues/2707
	resource, err := configService.GetKeptnResource(url.QueryEscape("job/config.yaml"))
	if err != nil {
		log.Printf("Could not find config for job-executor-service")
		return err
	}

	configuration, err := config.NewConfig(resource)
	if err != nil {
		log.Printf("Could not parse config: %s", err)
		return err
	}

	found, action := configuration.FindActionByName(actionName)
	if !found {
		return fmt.Errorf("no action found with name '%s'", actionName)
	}

	found, task := action.FindTaskByName(taskName)
	if !found {
		return fmt.Errorf("no task found with name '%s'", taskName)
	}

	for _, fileName := range task.Files {

		resource, err = configService.GetKeptnResource(url.QueryEscape(fileName))
		if err != nil {
			log.Printf("Could not find file %s for task %s", fileName, taskName)
			return err
		}

		// Our mount starts with /keptn
		dir := filepath.Join("/keptn", filepath.Dir(fileName))
		fullFilePath := filepath.Join("/keptn", fileName)

		err := fs.MkdirAll(dir, 0700)
		if err != nil {
			log.Printf("Could not create directory %s for file %s", dir, fileName)
			return err
		}

		file, err := fs.Create(fullFilePath)
		if err != nil {
			log.Printf("Could not create file %s", fileName)
			return err
		}

		_, err = file.Write(resource)
		defer func() {
			err = file.Close()
			if err != nil {
				log.Printf("Could not close file %s", file.Name())
			}
		}()

		if err != nil {
			log.Printf("Could not write to file %s", fileName)
			return err
		}

		log.Printf("Successfully moved file %s to %s", fileName, fullFilePath)
	}

	return nil
}
