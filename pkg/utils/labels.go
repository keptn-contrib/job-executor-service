package utils

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ReadAndValidateJobLabels reads the user defined labels from a yaml file and validates
func ReadAndValidateJobLabels(jobLabelsYamlPath string) (map[string]string, error) {

	jobLabelsYaml, err := ioutil.ReadFile(jobLabelsYamlPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s: %w", jobLabelsYamlPath, err)
	}

	var jobLabels map[string]string
	err = yaml.Unmarshal(jobLabelsYaml, &jobLabels)
	if err != nil {
		return nil, fmt.Errorf("unable to parse yaml content: %w", err)
	}

	errorList := validation.ValidateLabels(jobLabels, &field.Path{})
	if errorList != nil && len(errorList) > 0 {
		return nil, fmt.Errorf("validation faild: %w", errorList.ToAggregate())
	}

	return jobLabels, nil
}
