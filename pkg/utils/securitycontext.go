package utils

import (
	"encoding/json"
	"errors"
	v1 "k8s.io/api/core/v1"
	"keptn-contrib/job-executor-service/pkg/config"
	"log"
	"os"
)

// jobSecurityContextFilePath describes the path of the security config file that is defined in the deployment.yaml
const jobSecurityContextFilePath = "/config/job-defaultSecurityContext.json"

// podSecurityContextFilePath describes the path of the pod security config file that is defined in the deployment.yaml
const podSecurityContextFilePath = "/config/job-podSecurityContext.json"

// ReadDefaultJobSecurityContext reads the JSON file defined in jobSecurityContextFilePath and parses the output
// into a valid v1.SecurityContext data structure
func ReadDefaultJobSecurityContext() (*v1.SecurityContext, error) {
	jobSecurityContextBytes, err := os.ReadFile(jobSecurityContextFilePath)
	if err != nil {
		return nil, err
	}

	var defaultJobSecurityContext v1.SecurityContext
	err = json.Unmarshal(jobSecurityContextBytes, &defaultJobSecurityContext)
	if err != nil {
		return nil, err
	}

	return &defaultJobSecurityContext, nil
}

// ReadDefaultPodSecurityContext reads the JSON file defined in podSecurityContextFilePath and parses the output
// into a valid v1.PodSecurityContext data structure
func ReadDefaultPodSecurityContext() (*v1.PodSecurityContext, error) {
	podSecurityContextBytes, err := os.ReadFile(podSecurityContextFilePath)
	if err != nil {
		return nil, err
	}

	var defaultPodSecurityContext v1.PodSecurityContext
	err = json.Unmarshal(podSecurityContextBytes, &defaultPodSecurityContext)
	if err != nil {
		return nil, err
	}

	return &defaultPodSecurityContext, nil
}

// BuildSecurityContext builds the final v1.SecurityContext from the SecurityContext defined in the job configuration
// and the default job security context. The configs are merged in such a way that the configuration in the task will
// overwrite each entry that is defined.
func BuildSecurityContext(defaultSecurityContext *v1.SecurityContext, securityContext config.SecurityContext) *v1.SecurityContext {

	// DeepCopy the security context to prevent modifications to it
	finalSecurityContext := defaultSecurityContext.DeepCopy()

	// If present any property from the task security config will be used to overwrite the existing elements in the
	// default security context. For some properties like capabilities a new structure will be created and filled
	// accordingly

	if securityContext.Capabilities != nil {
		capabilityAddArray := make([]v1.Capability, len(securityContext.Capabilities.Add))
		for i, capability := range securityContext.Capabilities.Add {
			capabilityAddArray[i] = v1.Capability(capability)
		}

		capabilityDropArray := make([]v1.Capability, len(securityContext.Capabilities.Drop))
		for i, capability := range securityContext.Capabilities.Drop {
			capabilityDropArray[i] = v1.Capability(capability)
		}

		capabilities := v1.Capabilities{
			Add:  capabilityAddArray,
			Drop: capabilityDropArray,
		}

		finalSecurityContext.Capabilities = &capabilities
	}

	if securityContext.Privileged != nil {
		finalSecurityContext.Privileged = securityContext.Privileged
	}

	if securityContext.RunAsUser != nil {
		finalSecurityContext.RunAsUser = securityContext.RunAsUser
	}

	if securityContext.RunAsGroup != nil {
		finalSecurityContext.RunAsGroup = securityContext.RunAsGroup
	}

	if securityContext.RunAsNonRoot != nil {
		finalSecurityContext.RunAsNonRoot = securityContext.RunAsNonRoot
	}

	if securityContext.ReadOnlyRootFilesystem != nil {
		finalSecurityContext.ReadOnlyRootFilesystem = securityContext.ReadOnlyRootFilesystem
	}

	if securityContext.AllowPrivilegeEscalation != nil {
		finalSecurityContext.AllowPrivilegeEscalation = securityContext.AllowPrivilegeEscalation
	}

	if securityContext.ProcMount != nil {
		finalSecurityContext.ProcMount = (*v1.ProcMountType)(securityContext.ProcMount)
	}

	if securityContext.SeccompProfile != nil {
		seccompProfile := v1.SeccompProfile{
			Type:             v1.SeccompProfileType(*securityContext.SeccompProfile.Type),
			LocalhostProfile: securityContext.SeccompProfile.LocalhostProfile,
		}

		finalSecurityContext.SeccompProfile = &seccompProfile
	}

	return finalSecurityContext
}

type SecurityViolation int32

const (
	PrivilegedContainerViolation SecurityViolation = iota
	RunningAsRootViolation
)

// CheckJobSecurityContext checks the job security context for dangerous flags (like privileged=true) and returns
// these violations in an array. An empty array means that the configuration should be sound
func CheckJobSecurityContext(securityContext *v1.SecurityContext) []SecurityViolation {
	var violations []SecurityViolation

	// TODO: Extends the list of unwanted security flags:

	if securityContext.Privileged != nil && *securityContext.Privileged {
		violations = append(violations, PrivilegedContainerViolation)
	}

	if securityContext.RunAsNonRoot != nil && !*securityContext.RunAsNonRoot {
		violations = append(violations, RunningAsRootViolation)
	}

	return violations
}

// CheckPodSecurityContext checks the pod security context for dangerous flags and returns
// these violations in an array. An empty array means that the configuration should be sound
func CheckPodSecurityContext(podSecurityContext *v1.PodSecurityContext) []SecurityViolation {
	var violations []SecurityViolation

	// TODO: Extends the list of unwanted security flags:

	if podSecurityContext.RunAsNonRoot != nil && !*podSecurityContext.RunAsNonRoot {
		violations = append(violations, RunningAsRootViolation)
	}

	return violations
}

// VerifySecurityContext checks if the given security context of the pod & job is considered secure. If it isn't an
// error will be returned, while warnings are printed to the log
func VerifySecurityContext(podSecurityContext *v1.PodSecurityContext, jobSecurityContext *v1.SecurityContext, allowPrivilegedJobs bool) error {

	for _, violation := range CheckJobSecurityContext(jobSecurityContext) {
		switch violation {
		case PrivilegedContainerViolation:
			if allowPrivilegedJobs {
				log.Println("WARNING: Security context for jobs contains privileged=true")
			} else {
				return errors.New("privileged containers are not allowed")
			}
		case RunningAsRootViolation:
			log.Println("WARNING: Security context for jobs contains runAsNonRoot=true")
		}
	}

	for _, violation := range CheckPodSecurityContext(podSecurityContext) {
		switch violation {
		case RunningAsRootViolation:
			log.Println("WARNING: Pod security context for jobs contains runAsNonRoot=true!")
		}
	}

	return nil
}

// VerifySecurityConfiguration checks if the security context configuration that is present in the task is sound and
// does not contain flags that can be considered a security risk
func VerifySecurityConfiguration(config *config.Config, allowPrivilegedJobs bool) error {
	emptySecurityContext := new(v1.SecurityContext)
	emptyPodSecurityContext := new(v1.PodSecurityContext)

	for _, action := range config.Actions {
		for _, task := range action.Tasks {
			taskSecurityContext := BuildSecurityContext(emptySecurityContext, task.SecurityContext)

			err := VerifySecurityContext(emptyPodSecurityContext, taskSecurityContext, allowPrivilegedJobs)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
