package utils

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"keptn-contrib/job-executor-service/pkg/config"
	"testing"
)

func TestBuildSecurityContextWithEmptyConfig(t *testing.T) {
	emptyDefaultSecurityContext := new(v1.SecurityContext)
	var emptyTaskSecurityContext config.SecurityContext

	resultingSecurityContext := BuildSecurityContext(emptyDefaultSecurityContext, emptyTaskSecurityContext)

	assert.NotNil(t, resultingSecurityContext)
	assert.Equal(t, emptyDefaultSecurityContext, resultingSecurityContext)
}

func TestBuildSecurityContextWithOverwriting(t *testing.T) {
	falseValue := false
	user := int64(1000)

	defaultSecurityContext := v1.SecurityContext{
		Privileged:               &falseValue,
		RunAsNonRoot:             &falseValue,
		RunAsUser:                &user,
		RunAsGroup:               &user,
		AllowPrivilegeEscalation: &falseValue,
		ReadOnlyRootFilesystem:   &falseValue,
	}

	taskUser := int64(2000)
	trueValue := true

	taskSecurityContext := config.SecurityContext{
		RunAsUser:    &taskUser,
		RunAsGroup:   &taskUser,
		RunAsNonRoot: &trueValue,
	}

	resultingSecurityContext := BuildSecurityContext(&defaultSecurityContext, taskSecurityContext)
	assert.NotNil(t, resultingSecurityContext)

	expectedSecurityContext := v1.SecurityContext{
		Privileged:               &falseValue,
		RunAsNonRoot:             &trueValue,
		RunAsUser:                &taskUser,
		RunAsGroup:               &taskUser,
		AllowPrivilegeEscalation: &falseValue,
		ReadOnlyRootFilesystem:   &falseValue,
	}

	assert.Equal(t, &expectedSecurityContext, resultingSecurityContext)
}

func TestBuildSecurityContextWithComplexConfig(t *testing.T) {
	trueValue := true
	falseValue := false
	user := int64(1000)

	defaultSecurityContext := v1.SecurityContext{
		Privileged:               &falseValue,
		RunAsNonRoot:             &falseValue,
		RunAsUser:                &user,
		RunAsGroup:               &user,
		AllowPrivilegeEscalation: &trueValue,
		ReadOnlyRootFilesystem:   &falseValue,
	}

	taskUser := int64(2000)

	taskCapabilities := config.Capabilities{
		Add:  []config.Capability{"cap"},
		Drop: []config.Capability{"all"},
	}

	taskMountType := "Default"

	seccompProfileType := "Runtime/Default"
	taskSeccompProfile := config.SeccompProfile{
		Type: &seccompProfileType,
	}

	taskSecurityContext := config.SecurityContext{
		RunAsUser:                &taskUser,
		RunAsGroup:               &taskUser,
		RunAsNonRoot:             &trueValue,
		Capabilities:             &taskCapabilities,
		ProcMount:                &taskMountType,
		SeccompProfile:           &taskSeccompProfile,
		ReadOnlyRootFilesystem:   &falseValue,
		AllowPrivilegeEscalation: &falseValue,
	}

	resultingSecurityContext := BuildSecurityContext(&defaultSecurityContext, taskSecurityContext)
	assert.NotNil(t, resultingSecurityContext)

	expectedCapabilities := v1.Capabilities{
		Add:  []v1.Capability{"cap"},
		Drop: []v1.Capability{"all"},
	}

	expectedSeccompProfile := v1.SeccompProfile{
		Type: v1.SeccompProfileType(seccompProfileType),
	}

	expectedSecurityContext := v1.SecurityContext{
		Privileged:               &falseValue,
		RunAsNonRoot:             &trueValue,
		RunAsUser:                &taskUser,
		RunAsGroup:               &taskUser,
		AllowPrivilegeEscalation: &falseValue,
		ReadOnlyRootFilesystem:   &falseValue,
		Capabilities:             &expectedCapabilities,
		SeccompProfile:           &expectedSeccompProfile,
		ProcMount:                (*v1.ProcMountType)(&taskMountType),
	}

	assert.Equal(t, &expectedSecurityContext, resultingSecurityContext)
}

func TestCheckJobSecurityContextWithInsecureContext(t *testing.T) {
	insecureSecurityContext := new(v1.SecurityContext)
	privileged := true
	nonRoot := false
	user := int64(0)

	insecureSecurityContext.Privileged = &privileged
	insecureSecurityContext.RunAsNonRoot = &nonRoot
	insecureSecurityContext.RunAsUser = &user

	violations := CheckJobSecurityContext(insecureSecurityContext)

	assert.Len(t, violations, 2)
	assert.Contains(t, violations, PrivilegedContainerViolation)
	assert.Contains(t, violations, RunningAsRootViolation)
}

func TestCheckJobSecurityContextWithSecureContext(t *testing.T) {
	secureSecurityContext := new(v1.SecurityContext)
	privileged := false
	nonRoot := true
	user := int64(1000)

	secureSecurityContext.Privileged = &privileged
	secureSecurityContext.RunAsNonRoot = &nonRoot
	secureSecurityContext.RunAsUser = &user

	violations := CheckJobSecurityContext(secureSecurityContext)

	assert.Len(t, violations, 0)
}

func TestCheckPodSecurityContextWithInsecureContext(t *testing.T) {
	insecurePodSecurityContext := new(v1.PodSecurityContext)
	runAsRootNonRoot := false

	insecurePodSecurityContext.RunAsNonRoot = &runAsRootNonRoot

	violations := CheckPodSecurityContext(insecurePodSecurityContext)

	assert.Len(t, violations, 1)
	assert.Contains(t, violations, RunningAsRootViolation)
}

func TestCheckPodSecurityContextWithSecureContext(t *testing.T) {
	secureSecurityContext := new(v1.PodSecurityContext)
	privileged := true
	user := int64(1000)

	secureSecurityContext.RunAsNonRoot = &privileged
	secureSecurityContext.RunAsUser = &user
	secureSecurityContext.FSGroup = &user

	violations := CheckPodSecurityContext(secureSecurityContext)

	assert.Len(t, violations, 0)
}

func TestVerifySecurityContextWithSecureContext(t *testing.T) {
	falseValue := false
	user := int64(1000)

	defaultPodSecurityContext := v1.PodSecurityContext{
		RunAsGroup: &user,
		RunAsUser:  &user,
	}

	defaultSecurityContext := v1.SecurityContext{
		Privileged:               &falseValue,
		RunAsNonRoot:             &falseValue,
		RunAsUser:                &user,
		RunAsGroup:               &user,
		AllowPrivilegeEscalation: &falseValue,
		ReadOnlyRootFilesystem:   &falseValue,
	}

	err := VerifySecurityContext(&defaultPodSecurityContext, &defaultSecurityContext, false)
	assert.NoError(t, err)
}

func TestVerifySecurityContextWithInsecureContext(t *testing.T) {
	trueValue := true

	insecurePodSecurityContext := v1.PodSecurityContext{
		RunAsNonRoot: &trueValue,
	}
	privilegedSecurityContext := v1.SecurityContext{
		Privileged:   &trueValue,
		RunAsNonRoot: &trueValue,
	}

	err := VerifySecurityContext(&insecurePodSecurityContext, &privilegedSecurityContext, false)
	assert.Error(t, err)

	err = VerifySecurityContext(&insecurePodSecurityContext, &privilegedSecurityContext, true)
	assert.NoError(t, err)
}

func TestVerifySecurityConfigurationWithSecureContext(t *testing.T) {
	falseValue := false
	user := int64(1000)

	conf := config.Config{
		Actions: []config.Action{
			{
				Tasks: []config.Task{
					{
						SecurityContext: config.SecurityContext{
							RunAsUser:    &user,
							Privileged:   &falseValue,
							RunAsNonRoot: &falseValue,
						},
					},
				},
			},
		},
	}

	err := VerifySecurityConfiguration(&conf, false)
	assert.NoError(t, err)
}

func TestVerifySecurityConfigurationWithInsecureContext(t *testing.T) {
	trueValue := true
	user := int64(0)

	conf := config.Config{
		Actions: []config.Action{
			{
				Tasks: []config.Task{
					{
						SecurityContext: config.SecurityContext{
							RunAsUser:    &user,
							Privileged:   &trueValue,
							RunAsNonRoot: &trueValue,
						},
					},
				},
			},
		},
	}

	err := VerifySecurityConfiguration(&conf, false)
	assert.Error(t, err)

	err = VerifySecurityConfiguration(&conf, true)
	assert.NoError(t, err)
}
