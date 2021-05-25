package config

import (
	"encoding/json"
	"testing"

	"gotest.tools/assert"
)

const simpleConfig = `
actions:
  - name: "Run locust"
    events:
      - name: "sh.keptn.event.test.triggered"
        jsonpath:
          property: "$.test.teststrategy" 
          match: "health"
    tasks:
      - name: "Run locust smoke tests"
        files: 
          - locust/basic.py
          - locust/import.py
        image: "locustio/locust"
        cmd: "locust -f /keptn/locust/locustfile.py"
`

const complexConfig = `
actions:
  - name: "Run locust"
    events:
      - name: "sh.keptn.event.test.triggered"
        jsonpath:
          property: "$.test.teststrategy" 
          match: "locust"
    tasks:
      - name: "Run locust smoke tests"
        files: 
          - locust/basic.py
          - locust/import.py
        image: "locustio/locust"
        cmd: "locust -f /keptn/locust/locustfile.py"

  - name: "Run bash"
    events:
      - name: "sh.keptn.event.action.triggered"
        jsonpath: 
          property: "$.action.action"
          match: "hello"
      - name: "sh.keptn.event.action.triggered"
        jsonpath: 
          property: "$.action.action"
          match: "goodbye"
      - name: "sh.keptn.event.action.started"
    tasks:
      - name: "Run static world"
        image: "bash"
        cmd: "echo static"
      - name: "Run hello world"
        files: 
          - hello/hello-world.txt
        image: "bash"
        cmd: "cat /keptn/hello/heppo-world.txt | echo"
`

const testTriggeredEvent = `
{
  "data": {
    "deployment": {
      "deploymentNames": [
        "user_managed"
      ],
      "deploymentURIsLocal": [
        "https://keptn.sh"
      ],
      "deploymentURIsPublic": [
        ""
      ],
      "deploymentstrategy": "user_managed",
      "gitCommit": "eb5fc3d5253b1845d3d399c880c329374bbbb30e"
    },
    "message": "",
    "project": "sockshop",
    "stage": "dev",
    "service": "carts",
    "status": "succeeded",
    "test": {
      "teststrategy": "health"
    }
  },
  "id": "4fe1eed1-49e2-49a9-91af-a42c8b0f7811",
  "source": "shipyard-controller",
  "specversion": "1.0",
  "time": "2021-05-13T07:46:09.546Z",
  "type": "sh.keptn.event.test.triggered",
  "shkeptncontext": "138f7bf1-f027-42c4-b705-9033b5f5871e"
}`

func getActionEvent(subType string, action string) string {
	return `
{
  "type": "sh.keptn.event.action.` + subType + `",
  "specversion": "1.0",
  "source": "test-events",
  "id": "f2b878d3-03c0-4e8f-bc3f-454bc1b3d79b",
  "time": "2019-06-07T07:02:15.64489Z",
  "contenttype": "application/json",
  "shkeptncontext": "08735340-6f9e-4b32-97ff-3b6c292bc50i",
  "data": {
    "project": "sockshop",
    "stage": "dev",
    "service": "carts",
    "labels": {
      "testId": "4711",
      "buildId": "build-17",
      "owner": "JohnDoe"
    },
    "status": "succeeded",
    "result": "pass",
    "action": {
      "name": "run locust tests",
      "action": "` + action + `",
      "description": "so something as defined in remediation.yaml",
      "value": "1"
    },
    "problem": {}
  }
}`
}

func TestSimpleConfigUnmarshalling(t *testing.T) {

	config, err := NewConfig([]byte(simpleConfig))

	assert.NilError(t, err)
	assert.Equal(t, len(config.Actions), 1)
}

func TestComplexConfigUnmarshalling(t *testing.T) {

	config, err := NewConfig([]byte(complexConfig))

	assert.NilError(t, err)
	assert.Equal(t, len(config.Actions), 2)
}

func TestSimpleMatch(t *testing.T) {

	config, err := NewConfig([]byte(simpleConfig))
	assert.NilError(t, err)

	jsonEventData := interface{}(nil)
	err = json.Unmarshal([]byte(testTriggeredEvent), &jsonEventData)
	assert.NilError(t, err)

	data := jsonEventData.(map[string]interface{})["data"]
	found, action := config.IsEventMatch("sh.keptn.event.test.triggered", data)
	assert.Equal(t, found, true)
	assert.Equal(t, action.Events[0].Name, "sh.keptn.event.test.triggered")
}

func TestSimpleNoMatch(t *testing.T) {

	config, err := NewConfig([]byte(simpleConfig))
	assert.NilError(t, err)

	jsonEventData := interface{}(nil)
	err = json.Unmarshal([]byte(testTriggeredEvent), &jsonEventData)
	assert.NilError(t, err)

	data := jsonEventData.(map[string]interface{})["data"]
	found, _ := config.IsEventMatch("sh.keptn.event.action.triggered", data)
	assert.Equal(t, found, false)
}

func TestComplexMatch(t *testing.T) {

	config, err := NewConfig([]byte(complexConfig))
	assert.NilError(t, err)

	// sh.keptn.event.action.triggered - action: hello

	actionTriggeredEvent := getActionEvent("triggered", "hello")
	jsonEventData := interface{}(nil)
	err = json.Unmarshal([]byte(actionTriggeredEvent), &jsonEventData)
	assert.NilError(t, err)

	data := jsonEventData.(map[string]interface{})["data"]
	found, action := config.IsEventMatch("sh.keptn.event.action.triggered", data)
	assert.Equal(t, found, true)
	assert.Equal(t, action.Events[0].Name, "sh.keptn.event.action.triggered")

	// sh.keptn.event.action.triggered - action: goodbye

	actionTriggeredEvent = getActionEvent("triggered", "goodbye")
	jsonEventData = interface{}(nil)
	err = json.Unmarshal([]byte(actionTriggeredEvent), &jsonEventData)
	assert.NilError(t, err)

	data = jsonEventData.(map[string]interface{})["data"]
	found, action = config.IsEventMatch("sh.keptn.event.action.triggered", data)
	assert.Equal(t, found, true)
	assert.Equal(t, action.Events[1].Name, "sh.keptn.event.action.triggered")

	// sh.keptn.event.action.started - action: _

	actionTriggeredEvent = getActionEvent("started", "")
	jsonEventData = interface{}(nil)
	err = json.Unmarshal([]byte(actionTriggeredEvent), &jsonEventData)
	assert.NilError(t, err)

	data = jsonEventData.(map[string]interface{})["data"]
	found, action = config.IsEventMatch("sh.keptn.event.action.started", data)
	assert.Equal(t, found, true)
	assert.Equal(t, action.Events[2].Name, "sh.keptn.event.action.started")
}