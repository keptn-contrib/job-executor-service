package k8sutils

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// CreateResourceRequirements parses the given resource parameters and creates k8s resource requirements
func CreateResourceRequirements(resourceLimitsCPU, resourceLimitsMemory, resourceRequestsCPU, resourceRequestsMemory string) (*v1.ResourceRequirements, error) {
	resourceListLimits, err := createResourceList(resourceLimitsCPU, resourceLimitsMemory)
	if err != nil {
		return nil, fmt.Errorf("unable to parse resource limits requirement: %v", err.Error())
	}

	resourceListRequests, err := createResourceList(resourceRequestsCPU, resourceRequestsMemory)
	if err != nil {
		return nil, fmt.Errorf("unable to parse resource requests requirement: %v", err.Error())
	}

	resourceRequirements := v1.ResourceRequirements{
		Limits:   resourceListLimits,
		Requests: resourceListRequests,
	}

	return &resourceRequirements, nil
}

func createResourceList(cpu, memory string) (v1.ResourceList, error) {
	res := v1.ResourceList{}

	if cpu != "" {
		cpuQuantity, err := resource.ParseQuantity(cpu)
		if err != nil {
			return nil, fmt.Errorf("unable to parse cpu quantity '%v': %v", cpu, err.Error())
		}
		res[v1.ResourceCPU] = cpuQuantity
	}

	if memory != "" {
		memoryQuantity, err := resource.ParseQuantity(memory)
		if err != nil {
			return nil, fmt.Errorf("unable to parse memory quantity '%v': %v", memory, err.Error())
		}
		res[v1.ResourceMemory] = memoryQuantity
	}

	return res, nil
}
