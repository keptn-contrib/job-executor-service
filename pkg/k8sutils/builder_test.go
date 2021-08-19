package k8sutils

import (
	"gotest.tools/assert"
	"keptn-sandbox/job-executor-service/pkg/github/model"
	"log"
	"testing"
)

func TestBuilder(t *testing.T) {

	k8s := k8sImpl{}
	image, err := k8s.CreateImageBuilder("whatever", model.Step{}, "whatever")
	assert.NilError(t, err)
	log.Printf("%v", image)
}