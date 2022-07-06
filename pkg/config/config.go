package config

import (
	"fmt"
	"regexp"

	"gopkg.in/yaml.v2"

	"github.com/PaesslerAG/jsonpath"
)

const supportedAPIVersion = "v2"

// Config contains the configuration of the job-executor-service (job/config.yaml)
type Config struct {
	APIVersion *string  `yaml:"apiVersion"`
	Actions    []Action `yaml:"actions"`
}

// Action contains a action within the config which needs to be triggered
type Action struct {
	Name   string  `yaml:"name"`
	Events []Event `yaml:"events"`
	Tasks  []Task  `yaml:"tasks"`
	Silent bool    `yaml:"silent"`
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
	Name                    string            `yaml:"name"`
	Files                   []string          `yaml:"files"`
	Image                   string            `yaml:"image"`
	ImagePullPolicy         string            `yaml:"imagePullPolicy"`
	Cmd                     []string          `yaml:"cmd"`
	Args                    []string          `yaml:"args"`
	Env                     []Env             `yaml:"env"`
	Resources               *Resources        `yaml:"resources"`
	WorkingDir              string            `yaml:"workingDir"`
	MaxPollDuration         *int              `yaml:"maxPollDuration"`
	Namespace               string            `yaml:"namespace"`
	TTLSecondsAfterFinished *int32            `yaml:"ttlSecondsAfterFinished"`
	SecurityContext         SecurityContext   `yaml:"securityContext,omitempty"`
	ServiceAccount          *string           `yaml:"serviceAccount,omitempty"`
	Annotations             map[string]string `yaml:"annotations,omitempty"`
}

// Env value from the event which will be added as env to the job
type Env struct {
	Name       string `yaml:"name"`
	Value      string `yaml:"value"`
	ValueFrom  string `yaml:"valueFrom"`
	Formatting string `yaml:"as"`
}

// Resources defines the resource requirements of a task
type Resources struct {
	Limits   ResourceList `yaml:"limits"`
	Requests ResourceList `yaml:"requests"`
}

// ResourceList contains resource requirement keys
type ResourceList struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

// SecurityContext for the job container, it's a subset of the SecurityContext which is provided by Kubernetes
type SecurityContext struct {
	Capabilities             *Capabilities   `yaml:"capabilities,omitempty"`
	Privileged               *bool           `yaml:"privileged,omitempty"`
	RunAsUser                *int64          `yaml:"runAsUser,omitempty"`
	RunAsGroup               *int64          `yaml:"runAsGroup,omitempty"`
	RunAsNonRoot             *bool           `yaml:"runAsNonRoot,omitempty"`
	ReadOnlyRootFilesystem   *bool           `yaml:"readOnlyRootFilesystem,omitempty"`
	AllowPrivilegeEscalation *bool           `yaml:"allowPrivilegeEscalation,omitempty"`
	ProcMount                *string         `yaml:"procMount,omitempty"`
	SeccompProfile           *SeccompProfile `yaml:"seccompProfile,omitempty"`
}

// Capability represents the Capability string of the Kubernetes security context
type Capability string

// Capabilities contains the add and drop arrays for the Kubernetes security context
type Capabilities struct {
	Add  []Capability `yaml:"add,omitempty"`
	Drop []Capability `yaml:"drop,omitempty"`
}

// SELinuxOptions represents the Kubernetes SELinuxOptions in the security context
type SELinuxOptions struct {
	User  string `yaml:"user,omitempty"`
	Role  string `yaml:"role,omitempty"`
	Type  string `yaml:"type,omitempty"`
	Level string `yaml:"level,omitempty"`
}

// SeccompProfile represents the Kubernetes SeccompProfile in the security context
type SeccompProfile struct {
	Type             *string `yaml:"type"`
	LocalhostProfile *string `yaml:"localhostProfile,omitempty"`
}

// NewConfig creates a new configuration from the provided config file content
func NewConfig(yamlContent []byte) (*Config, error) {

	config := Config{}
	err := yaml.UnmarshalStrict(yamlContent, &config)

	if err != nil {
		return nil, err
	}

	if config.APIVersion == nil {
		return nil, fmt.Errorf("apiVersion must be specified")
	} else if *config.APIVersion != supportedAPIVersion {
		return nil, fmt.Errorf("apiVersion %v is not supported, use %v", *config.APIVersion, supportedAPIVersion)
	}

	return &config, nil
}

// IsEventMatch indicated whether a given event matches the config
func (c *Config) IsEventMatch(eventType string, jsonEventData interface{}) bool {

	for _, action := range c.Actions {
		if action.IsEventMatch(eventType, jsonEventData) {
			return true
		}
	}
	return false
}

// IsEventMatch indicated whether a given event matches the action
func (a *Action) IsEventMatch(eventType string, jsonEventData interface{}) bool {

	for _, event := range a.Events {
		// does the event type match with regex?
		matched, _ := regexp.MatchString(event.Name, eventType)
		if matched {

			// no JSONPath specified, just match event type
			if event.JSONPath.Property == "" {
				return true
			}

			value, err := jsonpath.Get(event.JSONPath.Property, jsonEventData)
			if err != nil {
				continue
			}

			if value == event.JSONPath.Match {
				return true
			}
		}
	}

	return false
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
