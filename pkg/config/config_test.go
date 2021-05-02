package config

import (
	"encoding/json"
	"testing"

	"gotest.tools/assert"
)

const simpleConfig = `
actions:
  - name: "Run locust"
    event: "sh.keptn.event.test.triggered"
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
    event: "sh.keptn.event.test.triggered"
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
    event: "sh.keptn.event.action.triggered"
    jsonpath: 
      property: "$.action.action"
      match: "hello"
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

const actionTriggeredEventData = `{
    "type": "sh.keptn.event.action.triggered",
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
        "action": "hello",
        "description": "so something as defined in remediation.yaml",
        "value" : "1"
      },
      "problem": {
      }
    }
  }`

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

	config, err := NewConfig([]byte(complexConfig))
	assert.NilError(t, err)

	jsonEventData := interface{}(nil)
	err = json.Unmarshal([]byte(actionTriggeredEventData), &jsonEventData)
	assert.NilError(t, err)

	data := jsonEventData.(map[string]interface{})["data"]
	found, action := config.IsEventMatch("sh.keptn.event.action.triggered", data)
	assert.Equal(t, found, true)
	assert.Equal(t, action.Event, "sh.keptn.event.action.triggered")
}

func TestSimpleNoMatch(t *testing.T) {

	config, err := NewConfig([]byte(simpleConfig))
	assert.NilError(t, err)

	jsonEventData := interface{}(nil)
	err = json.Unmarshal([]byte(actionTriggeredEventData), &jsonEventData)
	assert.NilError(t, err)

	data := jsonEventData.(map[string]interface{})["data"]
	found, _ := config.IsEventMatch("sh.keptn.event.action.triggered", data)
	assert.Equal(t, found, false)
}
