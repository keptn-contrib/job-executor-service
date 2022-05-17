package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReadAndValidateJobLabels_WithInvalidLabels(t *testing.T) {
	labels, err := ReadAndValidateJobLabels("./../test/data/invalid-user-defined-labels.yaml")
	assert.Error(t, err)
	assert.Nil(t, labels)
}

func TestReadAndValidateJobLabels_WithValidLabels(t *testing.T) {
	expectedLabels := map[string]string{
		"Label":           "Value",
		"Some_OtherLabel": "Value2",
	}

	labels, err := ReadAndValidateJobLabels("../../test/data/user-defined-labels.yaml")
	require.NoError(t, err)
	assert.Equal(t, expectedLabels, labels)
}
