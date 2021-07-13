package k8sutils

import (
	"context"
	"encoding/json"
	"keptn-sandbox/job-executor-service/pkg/config"
	"testing"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

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
    "stage": "dev",
    "service": "carts",
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
	_, err := k8s.prepareJobEnv(task, &eventData, eventAsInterface)
	assert.ErrorContains(t, err, "could not add env with name DEPLOYMENT_STRATEGY, unknown valueFrom ")
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
	jobEnv, err := k8s.prepareJobEnv(task, &eventData, eventAsInterface)
	assert.NilError(t, err)

	assert.Equal(t, jobEnv[0].Name, "HOST")
	assert.Equal(t, jobEnv[0].Value, "https://keptn.sh")

	assert.Equal(t, jobEnv[1].Name, "DEPLOYMENT_STRATEGY")
	assert.Equal(t, jobEnv[1].Value, "user_managed")

	assert.Equal(t, jobEnv[2].Name, "TEST_STRATEGY")
	assert.Equal(t, jobEnv[2].Value, "health")

	assert.Equal(t, jobEnv[3].Name, "KEPTN_PROJECT")
	assert.Equal(t, jobEnv[3].Value, "sockshop")

	assert.Equal(t, jobEnv[4].Name, "KEPTN_STAGE")
	assert.Equal(t, jobEnv[4].Value, "dev")

	assert.Equal(t, jobEnv[5].Name, "KEPTN_SERVICE")
	assert.Equal(t, jobEnv[5].Value, "carts")
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
	_, err := k8s.prepareJobEnv(task, &eventData, eventAsInterface)
	assert.ErrorContains(t, err, "unknown key undeploymentstrategy")
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
		namespace: "keptn",
	}

	secretName := "locust-sockshop-dev-carts"
	key1, value1 := "key1", "value1"
	key2, value2 := "key2", "value2"
	secretData := map[string][]byte{key1: []byte(value1), key2: []byte(value2)}
	k8sSecret := createK8sSecretObj(secretName, k8s.namespace, secretData)
	k8s.clientset.CoreV1().Secrets(k8s.namespace).Create(context.TODO(), k8sSecret, metav1.CreateOptions{})

	jobEnv, err := k8s.prepareJobEnv(task, &eventData, eventAsInterface)
	assert.NilError(t, err)

	// env from secrets can in in any order, sort them
	var orderedSecretEnv [2]*corev1.EnvVar
	for index, env := range jobEnv {
		if env.Name == key1 {
			orderedSecretEnv[0] = &jobEnv[index]
		} else if env.Name == key2 {
			orderedSecretEnv[1] = &jobEnv[index]
		}
	}

	assert.Assert(t, orderedSecretEnv[0] != nil, "env with key1 not present")
	assert.Equal(t, orderedSecretEnv[0].Name, key1)
	assert.Equal(t, orderedSecretEnv[0].ValueFrom.SecretKeyRef.Key, key1)
	assert.Equal(t, orderedSecretEnv[0].ValueFrom.SecretKeyRef.Name, secretName)

	assert.Assert(t, orderedSecretEnv[1] != nil, "env with key2 not present")
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
		namespace: "keptn",
	}
	_, err := k8s.prepareJobEnv(task, &eventData, eventAsInterface)
	assert.ErrorContains(t, err, "could not add env with name locust-sockshop-dev-carts, valueFrom secret: secrets \"locust-sockshop-dev-carts\" not found")
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
		namespace: "keptn",
	}

	jobEnv, err := k8s.prepareJobEnv(task, &eventData, eventAsInterface)
	assert.NilError(t, err)

	assert.Assert(t, len(jobEnv) == 4, "expected `jobEnv` to be 4, but was %d", len(jobEnv))
	assert.Equal(t, jobEnv[0].Name, envName)
	assert.Equal(t, jobEnv[0].Value, value)
}

func TestSetWorkingDir(t *testing.T) {
	jobName := "test-job-1"
	namespace := "keptn"
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
		namespace: "keptn",
	}

	err := k8s.CreateK8sJob(jobName, &config.Action{
		Name: jobName,
	}, config.Task{
		Name:       jobName,
		Image:      "alpine",
		Cmd:        []string{"ls"},
		WorkingDir: workingDir,
	}, &eventData, JobSettings{
		JobNamespace: namespace,
		DefaultResourceRequirements: &corev1.ResourceRequirements{
			Limits:   make(corev1.ResourceList),
			Requests: make(corev1.ResourceList),
		},
	}, "")

	assert.NilError(t, err)

	job, err := k8sClientSet.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
	assert.NilError(t, err)

	var container *corev1.Container

	for _, c := range job.Spec.Template.Spec.Containers {
		if c.Name == jobName {
			container = new(corev1.Container)
			*container = c
			break
		}
	}

	assert.Assert(t, container != nil, "No container called `%s` found", jobName)
	assert.Equal(t, container.WorkingDir, workingDir)

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
