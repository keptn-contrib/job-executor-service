package k8s

import (
	"didiladi/job-executor-service/pkg/config"
	"encoding/json"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"gotest.tools/assert"
	"testing"
)

const testTriggeredEvent = `
{
  "data": {
    "deployment": {
      "deploymentNames": [
        "user_managed"
      ],
      "deploymentURIsLocal": [
        "https://keptn.sh",
        "https://keptn2.sh"
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

func TestPrepareJobEnv(t *testing.T) {
	task := config.Task{
		Env: []config.Env{
			{
				Name:  "HOST",
				Value: "$.data.deployment.deploymentURIsLocal[0]",
			},
			{
				Name:  "DEPLOYMENT_STRATEGY",
				Value: "$.data.deployment.deploymentstrategy",
			},
			{
				Name:  "TEST_STRATEGY",
				Value: "$.data.test.teststrategy",
			},
		},
	}

	eventData := keptnv2.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
	}

	var eventAsInterface interface{}
	json.Unmarshal([]byte(testTriggeredEvent), &eventAsInterface)

	jobEnv, err := prepareJobEnv(task, &eventData, eventAsInterface)
	assert.NilError(t, err)

	assert.Equal(t, jobEnv[0].Name, "HOST")
	assert.Equal(t, jobEnv[0].Value, "https://keptn.sh")

	assert.Equal(t, jobEnv[1].Name, "DEPLOYMENT_STRATEGY")
	assert.Equal(t, jobEnv[1].Value, "user_managed")

	assert.Equal(t, jobEnv[2].Name, "TEST_STRATEGY")
	assert.Equal(t, jobEnv[2].Value, "health")

	assert.Equal(t, jobEnv[3].Name, "KEPTN_PROJECT")
	assert.Equal(t, jobEnv[3].Value, "sockshop")

	assert.Equal(t, jobEnv[4].Name, "KEPTN_STAGE")
	assert.Equal(t, jobEnv[4].Value, "dev")

	assert.Equal(t, jobEnv[5].Name, "KEPTN_SERVICE")
	assert.Equal(t, jobEnv[5].Value, "carts")
}

func TestPrepareJobEnvWithWrongJSONPath(t *testing.T) {
	task := config.Task{
		Env: []config.Env{
			{
				Name:  "DEPLOYMENT_STRATEGY",
				Value: "$.data.deployment.undeploymentstrategy",
			},
		},
	}

	eventData := keptnv2.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
	}

	var eventAsInterface interface{}
	json.Unmarshal([]byte(testTriggeredEvent), &eventAsInterface)

	_, err := prepareJobEnv(task, &eventData, eventAsInterface)
	assert.ErrorContains(t, err, "unknown key undeploymentstrategy")
}
