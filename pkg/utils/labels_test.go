package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConvertJSONMapToLabels(t *testing.T) {
	validLabels := `{"label1": "content1", "label2": "content2"}`
	excpectedLabels := map[string]string{
		"label1": "content1",
		"label2": "content2",
	}

	out, err := ConvertJSONMapToLabels(validLabels)
	require.NoError(t, err)
	assert.Equal(t, excpectedLabels, out)
}

func TestConvertJSONMapToLabelsWithInvalidLabels(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "Test_InvalidLabel",
			json: `{"0 INVALID\\\\LABEL [] KEY {-'#+'": "value"}`,
		},
		{
			name: "Test_InvalidKey",
			json: `{"key": "0 INVALID\\\\LABEL [] VALUE {-'#+'"}`,
		},
		{
			name: "Test_InvalidJson",
			json: `{"key": {"invalid": "json"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out, err := ConvertJSONMapToLabels(test.json)
			require.Error(t, err)
			require.Nil(t, out)
		})
	}
}
