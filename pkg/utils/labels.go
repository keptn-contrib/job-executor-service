package utils

import (
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ConvertJSONMapToLabels converts a json object to the label map and validates the contents
func ConvertJSONMapToLabels(JSON string) (map[string]string, error) {
	var labels map[string]string

	err := json.Unmarshal([]byte(JSON), &labels)
	if err != nil {
		return nil, fmt.Errorf("unable to parse JSON: %w", err)
	}

	errorList := validation.ValidateLabels(labels, &field.Path{})

	if errorList != nil && len(errorList) > 0 {
		validationError := errorList.ToAggregate().Error()
		return nil, fmt.Errorf("specified json is not valid: %s", validationError)
	}

	return labels, nil
}
