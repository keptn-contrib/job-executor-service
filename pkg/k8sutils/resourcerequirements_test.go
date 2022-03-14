package k8sutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
	assert.Equal(t, resourceRequirements.Limits.Cpu().String(), resourceLimitsCPU)
	assert.True(t, resourceRequirements.Limits.Memory().IsZero())
	assert.True(t, resourceRequirements.Requests.Cpu().IsZero())
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
	require.Error(t, err)
	assert.Contains(
		t, err.Error(),
		"unable to parse resource limits requirement: unable to parse cpu quantity '1KeinEiKummGeGimmeEinEi'",
	)
}
