package github

import (
	"fmt"
	"keptn-sandbox/job-executor-service/pkg/github/model"
	"log"
	"strings"
)

func PrepareArgs(with map[string]string, inputs map[string]model.Input, args []string) ([]string, error) {
	var filledArgs []string

	for inputKey, inputValue := range inputs {
		argKey := fmt.Sprintf("inputs.%s", inputKey)
		log.Printf("argKey: %v", argKey)

		for _, arg := range args {
			if strings.Contains(arg, argKey) {
				log.Printf("matched argKey: %v", argKey)

				argValue := inputValue.Default
				if withValue, ok := with[inputKey]; ok {
					argValue = withValue
				} else {
					if inputValue.Required {
						return nil, fmt.Errorf("required input '%s' not provided", inputKey)
					}
				}

				splittedArg := strings.Split(arg, "$")
				arg := strings.TrimSpace(splittedArg[0])
				filledArgs = append(filledArgs, arg)
				filledArgs = append(filledArgs, argValue)
			}
		}
	}

	return filledArgs, nil
}
