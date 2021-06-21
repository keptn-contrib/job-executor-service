package k8sutils

import (
	"gotest.tools/assert"
	"testing"
)

const (
	resourceLimitsCPU      = "1Ei"
	resourceLimitsMemory   = "512Mi"
	resourceRequestsCPU    = "50m"
	resourceRequestsMemory = "128Mi"
)

func TestCreateResourceRequirements_Valid(t *testing.T) {
	resourceRequirements, err := CreateResourceRequirements(
		resourceLimitsCPU,
		resourceLimitsMemory,
		resourceRequestsCPU,
		resourceRequestsMemory,
	)
	assert.NilError(t, err)
	assert.Equal(t, resourceRequirements.Limits.Cpu().String(), resourceLimitsCPU)
	assert.Equal(t, resourceRequirements.Limits.Memory().String(), resourceLimitsMemory)
	assert.Equal(t, resourceRequirements.Requests.Cpu().String(), resourceRequestsCPU)
	assert.Equal(t, resourceRequirements.Requests.Memory().String(), resourceRequestsMemory)
}

func TestCreateResourceRequirements_Partial(t *testing.T) {
	resourceRequirements, err := CreateResourceRequirements(
		resourceLimitsCPU,
		"",
		"",
		resourceRequestsMemory,
	)
	assert.NilError(t, err)
	assert.Equal(t, resourceRequirements.Limits.Cpu().String(), resourceLimitsCPU)
	assert.Assert(t, resourceRequirements.Limits.Memory().IsZero())
	assert.Assert(t, resourceRequirements.Requests.Cpu().IsZero())
	assert.Equal(t, resourceRequirements.Requests.Memory().String(), resourceRequestsMemory)
}

func TestCreateResourceRequirements_Invalid(t *testing.T) {
	// according to the k8s regex this is valid but the quantity suffix parsing afterwards should fail
	var resourceLimitsCPU = "1KeinEiKummGeGimmeEinEi"
	_, err := CreateResourceRequirements(
		resourceLimitsCPU,
		resourceLimitsMemory,
		resourceRequestsCPU,
		resourceRequestsMemory,
	)
	assert.ErrorContains(t, err, "unable to parse resource limits requirement: unable to parse cpu quantity '1KeinEiKummGeGimmeEinEi'")
}
