package github

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"keptn-sandbox/job-executor-service/pkg/github/model"
	"net/http"
)

func GetActionYaml(githubRepoName string) (error, *model.Action) {

	action := &model.Action{}

	err, actionAsString := readActionYamlFromGithub(githubRepoName)
	if err != nil {
		return err, action
	}

	err = yaml.Unmarshal(actionAsString, action)
	if err != nil {
		return err, action
	}

	return nil, action
}

func readActionYamlFromGithub(githubRepoName string) (error, []byte) {

	// e.g. https://raw.githubusercontent.com/aquasecurity/trivy-action/master/action.yaml
	response, err := http.Get("https://raw.githubusercontent.com/" + githubRepoName + "/master/action.yaml")

	if err != nil {
		return err, nil
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err, nil
		}
		return nil, bodyBytes
	}

	return fmt.Errorf("HTTP status code was: %v", response.StatusCode), nil
}
