package config

import (
	"gopkg.in/yaml.v2"
	"regexp"

	"github.com/PaesslerAG/jsonpath"
)

// Config contains the configuration of the job-executor-service (job/config.yaml)
type Config struct {
	Actions []Action `yaml:"actions"`
}

// Action contains a action within the config which needs to be triggered
type Action struct {
	Name   string  `yaml:"name"`
	Events []Event `yaml:"events"`
	Tasks  []Task  `yaml:"tasks"`
}

// Event defines a keptn event which determines if an Action should be triggered
type Event struct {
	Name     string   `yaml:"name"`
	JSONPath JSONPath `yaml:"jsonpath,omitempty"`
}

// JSONPath defines a filter for an Event
type JSONPath struct {
	Property string `yaml:"property"`
	Match    string `yaml:"match"`
}

// Task this is the actual task which can be triggered within an Action
type Task struct {
	Name  string   `yaml:"name"`
	Files []string `yaml:"files"`
	Image string   `yaml:"image"`
	Cmd   string   `yaml:"cmd"`
}

// NewConfig creates a new configuration from the provided config file content
func NewConfig(yamlContent []byte) (*Config, error) {

	config := Config{}
	err := yaml.Unmarshal(yamlContent, &config)

	return &config, err
}

// IsEventMatch indicated whether a given event matches the config
func (c *Config) IsEventMatch(eventType string, jsonEventData interface{}) (bool, *Action) {

	for _, action := range c.Actions {
		for _, event := range action.Events {
			// does the event type match with regex?
			matched, _ := regexp.MatchString(event.Name, eventType)
			if matched {

				// no JSONPath specified, just match event type
				if event.JSONPath.Property == "" {
					return true, &action
				}

				value, err := jsonpath.Get(event.JSONPath.Property, jsonEventData)
				if err != nil {
					continue
				}

				if value == event.JSONPath.Match {
					return true, &action
				}
			}
		}
	}
	return false, nil
}

// FindActionByName searches for a given Action by a provided name within the config
func (c *Config) FindActionByName(actionName string) (bool, *Action) {

	for _, action := range c.Actions {
		if actionName == action.Name {
			return true, &action
		}
	}
	return false, nil
}

// FindTaskByName searches for a given Task by a provided name within the config
func (a *Action) FindTaskByName(taskName string) (bool, *Task) {

	for _, task := range a.Tasks {
		if taskName == task.Name {
			return true, &task
		}
	}
	return false, nil
}
