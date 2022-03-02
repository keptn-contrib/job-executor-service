package k8sutils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"

	"keptn-contrib/job-executor-service/pkg/config"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

const testNamespace = "keptn"
const testTriggeredEvent = `
{
  "data": {
    "deployment": {
      "deploymentNames": [
        "user_managed"
      ],
      "deploymentURIsLocal": [
        "https://keptn.sh",
        "https://keptn2.sh"
      ],
      "deploymentURIsPublic": [
        ""
      ],
      "deploymentstrategy": "user_managed",
      "gitCommit": "eb5fc3d5253b1845d3d399c880c329374bbbb30e"
    },
    "message": "",
    "project": "sockshop",
    "service": "carts",
    "stage": "dev",
    "status": "succeeded",
    "test": {
      "teststrategy": "health"
    }
  },
  "id": "4fe1eed1-49e2-49a9-91af-a42c8b0f7811",
  "source": "shipyard-controller",
  "specversion": "1.0",
  "time": "2021-05-13T07:46:09.546Z",
  "type": "sh.keptn.event.test.triggered",
  "shkeptncontext": "138f7bf1-f027-42c4-b705-9033b5f5871e"
}`

func TestPrepareJobEnv_WithNoValueFrom(t *testing.T) {
	task := config.Task{
		Env: []config.Env{
			{
				Name:  "DEPLOYMENT_STRATEGY",
				Value: "$.data.deployment.undeploymentstrategy",
			},
		},
	}

	eventData := keptnv2.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
	}

	var eventAsInterface interface{}
	json.Unmarshal([]byte(testTriggeredEvent), &eventAsInterface)

	k8s := k8sImpl{}
	_, err := k8s.prepareJobEnv(task, &eventData, eventAsInterface, testNamespace)
	assert.EqualError(t, err, "could not add env with name DEPLOYMENT_STRATEGY, unknown valueFrom ")
}

func TestPrepareJobEnvFromEvent(t *testing.T) {
	task := config.Task{
		Env: []config.Env{
			{
				Name:      "HOST",
				Value:     "$.data.deployment.deploymentURIsLocal[0]",
				ValueFrom: "event",
			},
			{
				Name:      "DEPLOYMENT_STRATEGY",
				Value:     "$.data.deployment.deploymentstrategy",
				ValueFrom: "event",
			},
			{
				Name:      "TEST_STRATEGY",
				Value:     "$.data.test.teststrategy",
				ValueFrom: "event",
			},
			{
				Name:      "DATA_JSON",
				Value:     "$.data.deployment",
				ValueFrom: "event",
			},
			{
				Name:       "DATA_YAML",
				Value:      "$.data.deployment.deploymentURIsLocal",
				ValueFrom:  "event",
				Formatting: "yaml",
			},
		},
	}

	eventData := keptnv2.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
		Labels: map[string]string{
			"app-version":    "0.1.2",
			"build-datetime": "202202212056",
		},
	}

	var eventAsInterface interface{}
	json.Unmarshal([]byte(testTriggeredEvent), &eventAsInterface)

	k8s := k8sImpl{}
	jobEnv, err := k8s.prepareJobEnv(task, &eventData, eventAsInterface, testNamespace)
	require.NoError(t, err)

	testTriggeredEventJSON := `
{
  "deploymentNames": [
    "user_managed"
  ],
  "deploymentURIsLocal": [
    "https://keptn.sh",
    "https://keptn2.sh"
  ],
  "deploymentURIsPublic": [
    ""
  ],
  "deploymentstrategy": "user_managed",
  "gitCommit": "eb5fc3d5253b1845d3d399c880c329374bbbb30e"
}`
	testTriggeredEventJSON = strings.ReplaceAll(testTriggeredEventJSON, " ", "")
	testTriggeredEventJSON = strings.ReplaceAll(testTriggeredEventJSON, "\n", "")

	testTriggeredEventYaml := `- https://keptn.sh
- https://keptn2.sh
`
	assert.Equal(t, jobEnv[4].Name, "DATA_YAML")
	assert.Equal(t, jobEnv[4].Value, testTriggeredEventYaml)

	expectedEnv := []corev1.EnvVar{
		{Name: "HOST", Value: "https://keptn.sh"},
		{Name: "DEPLOYMENT_STRATEGY", Value: "user_managed"},
		{Name: "TEST_STRATEGY", Value: "health"},
		{Name: "DATA_JSON", Value: testTriggeredEventJSON},
		{Name: "DATA_YAML", Value: testTriggeredEventYaml},
		// built-ins:
		{Name: "KEPTN_PROJECT", Value: "sockshop"},
		{Name: "KEPTN_STAGE", Value: "dev"},
		{Name: "KEPTN_SERVICE", Value: "carts"},
		// labels:
		{Name: "LABELS_APP_VERSION", Value: "0.1.2"},
		{Name: "LABELS_BUILD_DATETIME", Value: "202202212056"},
	}

	assert.Subset(t, jobEnv, expectedEnv, "Prepared Job Environment does not contain the expected environment variables (or values)")
}

func TestPrepareJobEnvFromEvent_WithWrongJSONPath(t *testing.T) {
	task := config.Task{
		Env: []config.Env{
			{
				Name:      "DEPLOYMENT_STRATEGY",
				Value:     "$.data.deployment.undeploymentstrategy",
				ValueFrom: "event",
			},
		},
	}

	eventData := keptnv2.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
	}

	var eventAsInterface interface{}
	json.Unmarshal([]byte(testTriggeredEvent), &eventAsInterface)

	k8s := k8sImpl{}
	_, err := k8s.prepareJobEnv(task, &eventData, eventAsInterface, testNamespace)
	assert.Contains(t, err.Error(), "unknown key undeploymentstrategy")
}

func TestPrepareJobEnvFromSecret(t *testing.T) {
	task := config.Task{
		Env: []config.Env{
			{
				Name:      "locust-sockshop-dev-carts",
				ValueFrom: "secret",
			},
		},
	}

	eventData := keptnv2.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
	}

	var eventAsInterface interface{}
	json.Unmarshal([]byte(testTriggeredEvent), &eventAsInterface)

	k8s := k8sImpl{
		clientset: k8sfake.NewSimpleClientset(),
	}

	secretName := "locust-sockshop-dev-carts"
	key1, value1 := "key1", "value1"
	key2, value2 := "key2", "value2"
	secretData := map[string][]byte{key1: []byte(value1), key2: []byte(value2)}
	k8sSecret := createK8sSecretObj(secretName, testNamespace, secretData)
	k8s.clientset.CoreV1().Secrets(testNamespace).Create(context.TODO(), k8sSecret, metav1.CreateOptions{})

	jobEnv, err := k8s.prepareJobEnv(task, &eventData, eventAsInterface, testNamespace)
	require.NoError(t, err)

	// env from secrets can in in any order, sort them
	var orderedSecretEnv [2]*corev1.EnvVar
	for index, env := range jobEnv {
		if env.Name == key1 {
			orderedSecretEnv[0] = &jobEnv[index]
		} else if env.Name == key2 {
			orderedSecretEnv[1] = &jobEnv[index]
		}
	}

	require.NotNil(t, orderedSecretEnv[0], "env with key1 not present")
	assert.Equal(t, orderedSecretEnv[0].Name, key1)
	assert.Equal(t, orderedSecretEnv[0].ValueFrom.SecretKeyRef.Key, key1)
	assert.Equal(t, orderedSecretEnv[0].ValueFrom.SecretKeyRef.Name, secretName)

	require.NotNil(t, orderedSecretEnv[1], "env with key2 not present")
	assert.Equal(t, orderedSecretEnv[1].Name, key2)
	assert.Equal(t, orderedSecretEnv[1].ValueFrom.SecretKeyRef.Key, key2)
	assert.Equal(t, orderedSecretEnv[1].ValueFrom.SecretKeyRef.Name, secretName)
}

func TestPrepareJobEnvFromSecret_SecretNotFound(t *testing.T) {
	task := config.Task{
		Env: []config.Env{
			{
				Name:      "locust-sockshop-dev-carts",
				ValueFrom: "secret",
			},
		},
	}

	eventData := keptnv2.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
	}

	var eventAsInterface interface{}
	json.Unmarshal([]byte(testTriggeredEvent), &eventAsInterface)

	k8s := k8sImpl{
		clientset: k8sfake.NewSimpleClientset(),
	}
	_, err := k8s.prepareJobEnv(task, &eventData, eventAsInterface, testNamespace)
	assert.EqualError(t, err, "could not add env with name locust-sockshop-dev-carts, valueFrom secret: secrets \"locust-sockshop-dev-carts\" not found")
}

func TestPrepareJobEnvFromString(t *testing.T) {
	envName := "test-event"
	value := "test"
	task := config.Task{
		Env: []config.Env{
			{
				Name:      envName,
				Value:     value,
				ValueFrom: "string",
			},
		},
	}

	eventData := keptnv2.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
	}

	var eventAsInterface interface{}
	json.Unmarshal([]byte(testTriggeredEvent), &eventAsInterface)

	k8s := k8sImpl{
		clientset: k8sfake.NewSimpleClientset(),
	}

	jobEnv, err := k8s.prepareJobEnv(task, &eventData, eventAsInterface, testNamespace)
	require.NoError(t, err)

	assert.Equal(t, len(jobEnv), 4)
	assert.Equal(t, jobEnv[0].Name, envName)
	assert.Equal(t, jobEnv[0].Value, value)
}

func TestSetWorkingDir(t *testing.T) {
	jobName := "test-job-1"
	workingDir := "/test/dir"

	eventData := keptnv2.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
	}

	var eventAsInterface interface{}
	json.Unmarshal([]byte(testTriggeredEvent), &eventAsInterface)

	k8sClientSet := k8sfake.NewSimpleClientset()

	k8s := k8sImpl{
		clientset: k8sClientSet,
	}

	err := k8s.CreateK8sJob(jobName, &config.Action{
		Name: jobName,
	}, config.Task{
		Name:       jobName,
		Image:      "alpine",
		Cmd:        []string{"ls"},
		WorkingDir: workingDir,
	}, &eventData, JobSettings{
		JobNamespace: testNamespace,
		DefaultResourceRequirements: &corev1.ResourceRequirements{
			Limits:   make(corev1.ResourceList),
			Requests: make(corev1.ResourceList),
		},
	}, "", testNamespace)

	require.NoError(t, err)

	job, err := k8sClientSet.BatchV1().Jobs(testNamespace).Get(context.TODO(), jobName, metav1.GetOptions{})
	require.NoError(t, err)

	var container *corev1.Container

	for _, c := range job.Spec.Template.Spec.Containers {
		if c.Name == jobName {
			container = new(corev1.Container)
			*container = c
			break
		}
	}

	require.NotNil(t, container, nil, "No container called `%s` found", jobName)
	assert.Equal(t, container.WorkingDir, workingDir)

}

func TestSetCustomNamespace(t *testing.T) {
	jobName := "test-job-1"
	namespace := "my-custom-namespace"

	eventData := keptnv2.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
	}

	var eventAsInterface interface{}
	json.Unmarshal([]byte(testTriggeredEvent), &eventAsInterface)

	k8sClientSet := k8sfake.NewSimpleClientset()

	k8s := k8sImpl{
		clientset: k8sClientSet,
	}

	err := k8s.CreateK8sJob(jobName, &config.Action{
		Name: jobName,
	}, config.Task{
		Name:  jobName,
		Image: "alpine",
		Cmd:   []string{"ls"},
	}, &eventData, JobSettings{
		JobNamespace: namespace,
		DefaultResourceRequirements: &corev1.ResourceRequirements{
			Limits:   make(corev1.ResourceList),
			Requests: make(corev1.ResourceList),
		},
	}, "", namespace)

	require.NoError(t, err)

	job, err := k8sClientSet.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, job.Namespace, namespace, "Could not find container in namespace %s", namespace)
}

func TestSetEmptyNamespace(t *testing.T) {
	jobName := "test-job-1"
	namespace := ""

	eventData := keptnv2.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
	}

	var eventAsInterface interface{}
	json.Unmarshal([]byte(testTriggeredEvent), &eventAsInterface)

	k8sClientSet := k8sfake.NewSimpleClientset()

	k8s := k8sImpl{
		clientset: k8sClientSet,
	}

	err := k8s.CreateK8sJob(jobName, &config.Action{
		Name: jobName,
	}, config.Task{
		Name:  jobName,
		Image: "alpine",
		Cmd:   []string{"ls"},
	}, &eventData, JobSettings{
		JobNamespace: namespace,
		DefaultResourceRequirements: &corev1.ResourceRequirements{
			Limits:   make(corev1.ResourceList),
			Requests: make(corev1.ResourceList),
		},
	}, "", namespace)

	require.NoError(t, err)

	job, err := k8sClientSet.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, job.Namespace, namespace, "Could not find container in namespace %s", namespace)
}

func TestImagePullPolicy(t *testing.T) {

	tests := []struct {
		name               string
		inputPullPolicy    string
		expectedPullPolicy corev1.PullPolicy
	}{
		{
			name:               "No pull policy specified in input - server will assign a default value",
			inputPullPolicy:    "",
			expectedPullPolicy: corev1.PullPolicy(""),
		},
		{
			name:               `"Always" policy specified in input`,
			inputPullPolicy:    "Always",
			expectedPullPolicy: corev1.PullAlways,
		},
		{
			name:               `"Never" policy specified in input`,
			inputPullPolicy:    "Never",
			expectedPullPolicy: corev1.PullNever,
		},
		{
			name:               `"IfNotPresent"" policy specified in input`,
			inputPullPolicy:    "IfNotPresent",
			expectedPullPolicy: corev1.PullIfNotPresent,
		},
		{
			name: "`always` policy specified in input wrong case - works on client it will fail on" +
				" server",
			inputPullPolicy:    "always",
			expectedPullPolicy: corev1.PullPolicy("always"),
		},
	}
	for i, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {

				k8sClientSet := k8sfake.NewSimpleClientset()
				k8s := k8sImpl{clientset: k8sClientSet}

				jobName := fmt.Sprintf("ipp-job-%d", i)

				task := config.Task{
					Name:            fmt.Sprintf("noIppTask-%d", i),
					Image:           "someImage:someversion",
					Cmd:             []string{"someCmd"},
					ImagePullPolicy: test.inputPullPolicy,
				}

				eventData := keptnv2.EventData{
					Project: "keptnproject",
					Stage:   "dev",
					Service: "keptnservice",
				}

				namespace := "test-namespace"

				err := k8s.CreateK8sJob(
					jobName, &config.Action{Name: jobName}, task, &eventData, JobSettings{
						JobNamespace: namespace,
						DefaultResourceRequirements: &corev1.ResourceRequirements{
							Limits:   make(corev1.ResourceList),
							Requests: make(corev1.ResourceList),
						},
					}, "", namespace,
				)
				require.NoError(t, err)

				job, err := k8sClientSet.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
				require.NoError(t, err)

				assert.Equal(t, job.Spec.Template.Spec.Containers[0].ImagePullPolicy, test.expectedPullPolicy)
			},
		)
	}

}

func TestTTLSecondsAfterFinished(t *testing.T) {

	var DefaultTTLAfterFinished int32 = 21600
	var TenMinutesTTLAfterFinished int32 = 600
	var ImmediatedlyDeletableTTLAfterFinished int32 = 0

	tests := []struct {
		name                            string
		ttlSecondsAfterFinished         *int32
		expectedTTLSecondsAfterFinished int32
	}{
		{
			name:                            "No ttl specified in input - we should have the default job executor of 21600",
			ttlSecondsAfterFinished:         nil,
			expectedTTLSecondsAfterFinished: DefaultTTLAfterFinished,
		},
		{
			name:                            "10 mins ttl specified in input",
			ttlSecondsAfterFinished:         &TenMinutesTTLAfterFinished,
			expectedTTLSecondsAfterFinished: TenMinutesTTLAfterFinished,
		},
		{
			name:                            "0 seconds ttl specified in input - job eligible for deletion immediately",
			ttlSecondsAfterFinished:         &ImmediatedlyDeletableTTLAfterFinished,
			expectedTTLSecondsAfterFinished: ImmediatedlyDeletableTTLAfterFinished,
		},
	}
	for i, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {

				k8sClientSet := k8sfake.NewSimpleClientset()
				k8s := k8sImpl{clientset: k8sClientSet}

				jobName := fmt.Sprintf("ipp-job-%d", i)

				task := config.Task{
					Name:                    fmt.Sprintf("noIppTask-%d", i),
					Image:                   "someImage:someversion",
					Cmd:                     []string{"someCmd"},
					TTLSecondsAfterFinished: test.ttlSecondsAfterFinished,
				}

				eventData := keptnv2.EventData{
					Project: "keptnproject",
					Stage:   "dev",
					Service: "keptnservice",
				}

				namespace := "test-namespace"

				err := k8s.CreateK8sJob(
					jobName, &config.Action{Name: jobName}, task, &eventData, JobSettings{
						JobNamespace: namespace,
						DefaultResourceRequirements: &corev1.ResourceRequirements{
							Limits:   make(corev1.ResourceList),
							Requests: make(corev1.ResourceList),
						},
					}, "", namespace,
				)
				require.NoError(t, err)

				job, err := k8sClientSet.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
				require.NoError(t, err)

				require.NotNil(t, job.Spec.TTLSecondsAfterFinished)
				assert.Equal(t, *job.Spec.TTLSecondsAfterFinished, test.expectedTTLSecondsAfterFinished)
			},
		)
	}

}

func createK8sSecretObj(name string, namespace string, data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
		Type: "Opaque",
	}
}
