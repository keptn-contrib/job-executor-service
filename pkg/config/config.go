package config

import (
	"gopkg.in/yaml.v2"

	"github.com/PaesslerAG/jsonpath"
)

type Config struct {
	Configuration Configuration `yaml:"configuration"`
	Actions       []Action      `yaml:"actions"`
}

type Action struct {
	Name     string   `yaml:"name"`
	Event    string   `yaml:"event"`
	JsonPath JsonPath `yaml:"jsonpath"`
	Tasks    []Task   `yaml:"tasks"`
}

type JsonPath struct {
	Property string `yaml:"property"`
	Match    string `yaml:"match"`
}

type Configuration struct {
	ConfigurationService ConfigurationService `yaml:"configurationService"`
}

type ConfigurationService struct {
	Url                   string `yaml:"url"`
	CredentialsSecretName string `yaml:"credentialsSecretName"`
}

type Task struct {
	Name  string   `yaml:"name"`
	Files []string `yaml:"files"`
	Image string   `yaml:"image"`
	Cmd   string   `yaml:"cmd"`
}

func NewConfig(yamlContent []byte) (*Config, error) {

	config := Config{}
	err := yaml.Unmarshal(yamlContent, &config)

	return &config, err
}

func (c *Config) IsEventMatch(event string, jsonEventData interface{}) (bool, *Action) {

	for _, action := range c.Actions {

		// does the event type match?
		if action.Event == event {

			value, err := jsonpath.Get(action.JsonPath.Property, jsonEventData)
			if err != nil {
				continue
			}

			if value == action.JsonPath.Match {
				return true, &action
			}
		}
	}
	return false, nil
}

func (c *Config) FindActionByName(actionName string) (bool, *Action) {

	for _, action := range c.Actions {
		if actionName == action.Name {
			return true, &action
		}
	}
	return false, nil
}

func (a *Action) FindTaskByName(taskName string) (bool, *Task) {

	for _, task := range a.Tasks {
		if taskName == task.Name {
			return true, &task
		}
	}
	return false, nil
}
