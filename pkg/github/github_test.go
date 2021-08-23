package github

import (
	"gotest.tools/assert"
	"keptn-sandbox/job-executor-service/pkg/github/model"
	"log"
	"testing"
)

func TestPrepareArgs(t *testing.T) {
	with := map[string]string{"scan-type": "banana", "format": "cucumber"}
	inputs := map[string]model.Input{
		"scan-type": {
			Required: false,
			Default:  "",
		},
		"format": {
			Required: false,
			Default:  "",
		},
		"template": {
			Required: false,
			Default:  "table",
		},
	}
	args := []string{"-a ${{ inputs.scan-type }}", "-b ${{ inputs.format }}", "-c ${{ inputs.template }}"}
	k8sArgs, err := PrepareArgs(with, inputs, args)
	assert.NilError(t, err)
	log.Printf("%v", k8sArgs)
}

func TestPrepareArgs_RequiredInput(t *testing.T) {
	with := map[string]string{}
	inputs := map[string]model.Input{
		"scan-type": {
			Required: true,
			Default:  "",
		},
	}
	args := []string{"-a ${{ inputs.scan-type }}"}
	_, err := PrepareArgs(with, inputs, args)
	assert.Error(t, err, "required input 'scan-type' not provided")
}
