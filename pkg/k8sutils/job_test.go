package k8sutils

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/batch/v1"

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

	k8s := K8sImpl{}
	_, err := k8s.prepareJobEnv(&task, &eventData, eventAsInterface, testNamespace)
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

	k8s := K8sImpl{}
	jobEnv, err := k8s.prepareJobEnv(&task, &eventData, eventAsInterface, testNamespace)
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

	assert.Subset(
		t, jobEnv, expectedEnv,
		"Prepared Job Environment does not contain the expected environment variables (or values)",
	)
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

	k8s := K8sImpl{}
	_, err := k8s.prepareJobEnv(&task, &eventData, eventAsInterface, testNamespace)
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

	k8s := K8sImpl{
		clientset: k8sfake.NewSimpleClientset(),
	}

	secretName := "locust-sockshop-dev-carts"
	key1, value1 := "key1", "value1"
	key2, value2 := "key2", "value2"
	secretData := map[string][]byte{key1: []byte(value1), key2: []byte(value2)}
	k8sSecret := createK8sSecretObj(secretName, testNamespace, secretData)
	k8s.clientset.CoreV1().Secrets(testNamespace).Create(context.TODO(), k8sSecret, metav1.CreateOptions{})

	jobEnv, err := k8s.prepareJobEnv(&task, &eventData, eventAsInterface, testNamespace)
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

	k8s := K8sImpl{
		clientset: k8sfake.NewSimpleClientset(),
	}
	_, err := k8s.prepareJobEnv(&task, &eventData, eventAsInterface, testNamespace)
	assert.EqualError(
		t, err,
		"could not add env with name locust-sockshop-dev-carts, valueFrom secret: secrets \"locust-sockshop-dev-carts\" not found",
	)
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

	k8s := K8sImpl{
		clientset: k8sfake.NewSimpleClientset(),
	}

	jobEnv, err := k8s.prepareJobEnv(&task, &eventData, eventAsInterface, testNamespace)
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

	k8s := K8sImpl{
		clientset: k8sClientSet,
	}

	err := k8s.CreateK8sJob(
		jobName,
		JobDetails{
			Action: &config.Action{
				Name: jobName,
			},
			Task: &config.Task{
				Name:       jobName,
				Image:      "alpine",
				Cmd:        []string{"ls"},
				WorkingDir: workingDir,
			},
			ActionIndex:   0,
			TaskIndex:     0,
			JobConfigHash: "",
		},
		&eventData, JobSettings{
			JobNamespace: testNamespace,
			DefaultResourceRequirements: &corev1.ResourceRequirements{
				Limits:   make(corev1.ResourceList),
				Requests: make(corev1.ResourceList),
			},
			DefaultPodSecurityContext: new(corev1.PodSecurityContext),
			DefaultSecurityContext:    new(corev1.SecurityContext),
		}, eventAsInterface, testNamespace,
	)

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

	k8s := K8sImpl{
		clientset: k8sClientSet,
	}

	err := k8s.CreateK8sJob(
		jobName,
		JobDetails{
			Action: &config.Action{
				Name: jobName,
			},
			Task: &config.Task{
				Name:  jobName,
				Image: "alpine",
				Cmd:   []string{"ls"},
			},
			ActionIndex:   0,
			TaskIndex:     0,
			JobConfigHash: "",
		},
		&eventData, JobSettings{
			JobNamespace: namespace,
			DefaultResourceRequirements: &corev1.ResourceRequirements{
				Limits:   make(corev1.ResourceList),
				Requests: make(corev1.ResourceList),
			},
			DefaultPodSecurityContext: new(corev1.PodSecurityContext),
			DefaultSecurityContext:    new(corev1.SecurityContext),
		}, eventAsInterface, namespace,
	)

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

	k8s := K8sImpl{
		clientset: k8sClientSet,
	}

	err := k8s.CreateK8sJob(
		jobName,
		JobDetails{
			Action: &config.Action{
				Name: jobName,
			},
			Task: &config.Task{
				Name:  jobName,
				Image: "alpine",
				Cmd:   []string{"ls"},
			},
			ActionIndex:   0,
			TaskIndex:     0,
			JobConfigHash: "",
		},
		&eventData, JobSettings{
			JobNamespace: namespace,
			DefaultResourceRequirements: &corev1.ResourceRequirements{
				Limits:   make(corev1.ResourceList),
				Requests: make(corev1.ResourceList),
			},
			DefaultPodSecurityContext: new(corev1.PodSecurityContext),
			DefaultSecurityContext:    new(corev1.SecurityContext),
		}, eventAsInterface, namespace,
	)

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
				k8s := K8sImpl{clientset: k8sClientSet}

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

				var eventAsInterface interface{}
				json.Unmarshal([]byte(testTriggeredEvent), &eventAsInterface)

				err := k8s.CreateK8sJob(
					jobName,
					JobDetails{
						Action: &config.Action{
							Name: jobName,
						},
						Task:          &task,
						ActionIndex:   0,
						TaskIndex:     0,
						JobConfigHash: "",
					}, &eventData, JobSettings{
						JobNamespace: namespace,
						DefaultResourceRequirements: &corev1.ResourceRequirements{
							Limits:   make(corev1.ResourceList),
							Requests: make(corev1.ResourceList),
						},
						DefaultPodSecurityContext: new(corev1.PodSecurityContext),
						DefaultSecurityContext:    new(corev1.SecurityContext),
					}, eventAsInterface, namespace,
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
			expectedTTLSecondsAfterFinished: minTTLSecondsAfterFinished,
		},
	}
	for i, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {

				k8sClientSet := k8sfake.NewSimpleClientset()
				k8s := K8sImpl{clientset: k8sClientSet}

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

				var eventAsInterface interface{}
				json.Unmarshal([]byte(testTriggeredEvent), &eventAsInterface)

				err := k8s.CreateK8sJob(
					jobName,
					JobDetails{
						Action: &config.Action{
							Name: jobName,
						},
						Task:          &task,
						ActionIndex:   0,
						TaskIndex:     0,
						JobConfigHash: "",
					}, &eventData, JobSettings{
						JobNamespace: namespace,
						DefaultResourceRequirements: &corev1.ResourceRequirements{
							Limits:   make(corev1.ResourceList),
							Requests: make(corev1.ResourceList),
						},
						DefaultPodSecurityContext: new(corev1.PodSecurityContext),
						DefaultSecurityContext:    new(corev1.SecurityContext),
					}, eventAsInterface, namespace,
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

func TestAwaitK8sJobDoneHappyPath(t *testing.T) {
	k8sClientSet := k8sfake.NewSimpleClientset()
	k8s := K8sImpl{clientset: k8sClientSet}

	jobName := "happy-path-job"
	job := v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: v1.JobSpec{},
		Status: v1.JobStatus{
			Conditions: []v1.JobCondition{
				{
					Type:   v1.JobComplete,
					Status: corev1.ConditionTrue,
					Reason: "Job has completed successfully!",
				},
			},
		},
	}
	namespace := "happy-path-ns"
	k8sClientSet.BatchV1().Jobs(namespace).Create(
		context.Background(), &job, metav1.CreateOptions{},
	)

	err := k8s.AwaitK8sJobDone(jobName, 1*time.Second, 50*time.Millisecond, namespace)

	assert.NoError(t, err)
}

func TestAwaitK8sJobDoneErrorJobFailed(t *testing.T) {
	k8sClientSet := k8sfake.NewSimpleClientset()
	k8s := K8sImpl{clientset: k8sClientSet}

	jobName := "failed-job"
	failedReason := "Job has gone horribly wrong!"
	failureMessage := "there has been a problem somewhere"
	job := v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: v1.JobSpec{},
		Status: v1.JobStatus{
			Conditions: []v1.JobCondition{
				{
					Type:    v1.JobFailed,
					Status:  corev1.ConditionTrue,
					Reason:  failedReason,
					Message: failureMessage,
				},
			},
		},
	}
	namespace := "job-pain-and-misery-ns"
	k8sClientSet.BatchV1().Jobs(namespace).Create(
		context.Background(), &job, metav1.CreateOptions{},
	)

	err := k8s.AwaitK8sJobDone(jobName, 1*time.Second, 50*time.Millisecond, namespace)

	require.Error(t, err)

	assert.ErrorContains(
		t, err,
		fmt.Sprintf(
			"job %s failed. Reason: %s, Message: %s", jobName, failedReason,
			failureMessage,
		),
	)
}

var Deadline30Sec = int64(30)
var ExpectedDeadline30Sec = int64(30)

func TestCreateJobTaskDeadlineSeconds(t *testing.T) {
	k8sClientSet := k8sfake.NewSimpleClientset()
	k8s := K8sImpl{clientset: k8sClientSet}

	tests := []struct {
		name                          string
		taskDeadlineSeconds           *int64
		expectedActiveDeadlineSeconds *int64
	}{
		{
			name:                          "No deadline specified, no limit set in job",
			taskDeadlineSeconds:           nil,
			expectedActiveDeadlineSeconds: nil,
		},
		{
			name:                          "30sec deadline",
			taskDeadlineSeconds:           &Deadline30Sec,
			expectedActiveDeadlineSeconds: &ExpectedDeadline30Sec,
		},
	}

	for i, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				jobName := fmt.Sprintf("tds-job-%d", i)
				task := config.Task{
					Name:  fmt.Sprintf("TdsTask-%d", i),
					Image: "someImage:someversion",
					Cmd:   []string{"someCmd"},
				}

				eventData := keptnv2.EventData{
					Project: "keptnproject",
					Stage:   "dev",
					Service: "keptnservice",
				}

				namespace := "test-namespace"

				jobSettings := JobSettings{
					JobNamespace: namespace,
					DefaultResourceRequirements: &corev1.ResourceRequirements{
						Limits:   make(corev1.ResourceList),
						Requests: make(corev1.ResourceList),
					},
					DefaultPodSecurityContext: new(corev1.PodSecurityContext),
					DefaultSecurityContext:    new(corev1.SecurityContext),
					TaskDeadlineSeconds:       test.taskDeadlineSeconds,
				}
				err := k8s.CreateK8sJob(
					jobName, &config.Action{Name: fmt.Sprintf("test-action-%d", i)}, task, &eventData, jobSettings, "",
					namespace,
				)

				require.NoError(t, err, "Error creating test job")

				job, err := k8sClientSet.BatchV1().Jobs(namespace).Get(
					context.Background(), jobName, metav1.GetOptions{},
				)

				require.NoError(t, err, "Error retrieving created test job")
				assert.Equal(t, test.expectedActiveDeadlineSeconds, job.Spec.ActiveDeadlineSeconds)
			},
		)
	}
}

func TestAwaitK8sJobDoneErrorJobSuspended(t *testing.T) {
	k8sClientSet := k8sfake.NewSimpleClientset()
	k8s := K8sImpl{clientset: k8sClientSet}

	jobName := "suspender-job"
	suspendedReason := "Job has been suspended"
	suspendedMessage := "some admin suspended your job"
	job := v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: v1.JobSpec{},
		Status: v1.JobStatus{
			Conditions: []v1.JobCondition{
				{
					Type:    v1.JobSuspended,
					Status:  corev1.ConditionTrue,
					Reason:  suspendedReason,
					Message: suspendedMessage,
				},
			},
		},
	}
	namespace := "job-suspended-ns"
	k8sClientSet.BatchV1().Jobs(namespace).Create(
		context.Background(), &job, metav1.CreateOptions{},
	)

	err := k8s.AwaitK8sJobDone(jobName, 1*time.Second, 50*time.Millisecond, namespace)

	require.Error(t, err)

	assert.ErrorContains(
		t, err,
		fmt.Sprintf(
			"job %s was suspended. Reason: %s, Message: %s", jobName,
			suspendedReason,
			suspendedMessage,
		),
	)
}

func TestAwaitK8sJobDoneErrorNeverComplete(t *testing.T) {
	k8sClientSet := k8sfake.NewSimpleClientset()
	k8s := K8sImpl{clientset: k8sClientSet}

	jobName := "looong-running-job"
	job := v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec:   v1.JobSpec{},
		Status: v1.JobStatus{},
	}
	namespace := "never-ending-jobs-land"
	k8sClientSet.BatchV1().Jobs(namespace).Create(
		context.Background(), &job, metav1.CreateOptions{},
	)

	err := k8s.AwaitK8sJobDone(jobName, 500*time.Millisecond, 50*time.Millisecond, namespace)

	require.Error(t, err)

	assert.ErrorContains(
		t, err, fmt.Sprintf(
			"polling for job %s timing out after", jobName,
		),
	)

	assert.ErrorIs(t, err, ErrMaxPollTimeExceeded)
}

func TestAwaitK8sJobExceededDeadline(t *testing.T) {
	k8sClientSet := k8sfake.NewSimpleClientset()
	k8s := K8sImpl{clientset: k8sClientSet}

	jobName := "deadline-exceeding-job"
	job := v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: v1.JobSpec{},
		Status: v1.JobStatus{
			Conditions: []v1.JobCondition{
				{
					Type:    v1.JobFailed,
					Status:  corev1.ConditionTrue,
					Reason:  reasonJobDeadlineExceeded,
					Message: "Job exceeded deadline",
				},
			},
		},
	}
	namespace := "tight-job-runtimes-only"
	k8sClientSet.BatchV1().Jobs(namespace).Create(
		context.Background(), &job, metav1.CreateOptions{},
	)

	err := k8s.AwaitK8sJobDone(jobName, 500*time.Millisecond, 50*time.Millisecond, namespace)

	require.Error(t, err)

	assert.ErrorContains(
		t, err, fmt.Sprintf("job %s failed:", jobName),
	)

	assert.ErrorIs(t, err, ErrTaskDeadlineExceeded)
}

func TestAwaitK8sJobDoneSuccessAfterPolling(t *testing.T) {
	k8sClientSet := k8sfake.NewSimpleClientset()
	k8s := K8sImpl{clientset: k8sClientSet}

	jobName := "slow-running-job"
	job := v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: v1.JobSpec{},
		Status: v1.JobStatus{
			Conditions: []v1.JobCondition{},
		},
	}
	namespace := "slow-ns"
	k8sClientSet.BatchV1().Jobs(namespace).Create(
		context.Background(), &job, metav1.CreateOptions{},
	)

	jobCompletionTimer := time.NewTimer(500 * time.Millisecond)

	go func() {
		select {
		case <-jobCompletionTimer.C:
			job.Status.Conditions = []v1.JobCondition{
				{
					Type:    v1.JobComplete,
					Status:  corev1.ConditionTrue,
					Reason:  "Job completed eventually!",
					Message: "Hooray!",
				},
			}
			k8sClientSet.BatchV1().Jobs(namespace).Update(context.Background(), &job, metav1.UpdateOptions{})
		}
	}()

	err := k8s.AwaitK8sJobDone(jobName, 2*time.Second, 50*time.Millisecond, namespace)

	require.NoError(t, err)
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

func TestCreateK8sJobContainsCorrectLabels(t *testing.T) {
	k8sClientSet := k8sfake.NewSimpleClientset()
	k8s := K8sImpl{clientset: k8sClientSet}

	eventData := keptnv2.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
	}

	var eventAsInterfaceWithoutGitCommitID map[string]interface{}
	err := json.Unmarshal([]byte(testTriggeredEvent), &eventAsInterfaceWithoutGitCommitID)
	require.NoError(t, err)

	var eventAsInterfaceWithGitCommitID map[string]interface{}
	data, err := ioutil.ReadFile("../../test/events/test.triggered.with-gitcommitid.json")
	require.NoError(t, err)

	err = json.Unmarshal(data, &eventAsInterfaceWithGitCommitID)
	require.NoError(t, err)

	namespace := testNamespace

	tests := []struct {
		name               string
		actionName         string
		taskName           string
		expectedActionName string
		expectedTaskName   string
		event              map[string]interface{}
	}{
		{
			name:               "Test normal Event with gitcommitid",
			actionName:         "some_action",
			taskName:           "task",
			event:              eventAsInterfaceWithGitCommitID,
			expectedActionName: "some_action",
			expectedTaskName:   "task",
		},
		{
			name:               "Test normal Event without gitcommitid",
			actionName:         "some_action",
			taskName:           "task",
			event:              eventAsInterfaceWithoutGitCommitID,
			expectedActionName: "some_action",
			expectedTaskName:   "task",
		},
		{
			name:               "Test non k8s compatible action name",
			actionName:         "Some fancy action name ...",
			taskName:           "--NameThat's not Compatible with k8s []{}///// a....",
			event:              eventAsInterfaceWithGitCommitID,
			expectedActionName: "Some_fancy_action_name",
			expectedTaskName:   "NameThat_s_not_Compatible_with_k8s_a",
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			jobName := "some-job-name-" + strconv.Itoa(i)
			err = k8s.CreateK8sJob(
				jobName,
				JobDetails{
					Action: &config.Action{
						Name: test.actionName,
					},
					Task: &config.Task{
						Name:  test.taskName,
						Image: "alpine",
						Cmd:   []string{"ls"},
					},
					ActionIndex:   0,
					TaskIndex:     0,
					JobConfigHash: "",
				},
				&eventData,
				JobSettings{
					JobNamespace: namespace,
					DefaultResourceRequirements: &corev1.ResourceRequirements{
						Limits:   make(corev1.ResourceList),
						Requests: make(corev1.ResourceList),
					},
					DefaultPodSecurityContext: new(corev1.PodSecurityContext),
					DefaultSecurityContext:    new(corev1.SecurityContext),
				},
				test.event,
				namespace,
			)
			require.NoError(t, err)

			job, err := k8sClientSet.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
			require.NoError(t, err)

			expectedLabels := map[string]string{
				"app.kubernetes.io/managed-by": "job-executor-service",
				"keptn.sh/context":             test.event["shkeptncontext"].(string),
				"keptn.sh/ceid":                test.event["id"].(string),
				"keptn.sh/commitid":            "",
				"keptn.sh/jes-action":          test.expectedActionName,
				"keptn.sh/jes-task":            test.expectedTaskName,
				"keptn.sh/job-confighash":      "",
				"keptn.sh/jes-action-index":    "0",
				"keptn.sh/jes-task-index":      "0",
			}

			if test.event["gitcommitid"] != nil {
				expectedLabels["keptn.sh/commitid"] = test.event["gitcommitid"].(string)
			}

			assert.Equal(t, expectedLabels, job.Labels)
		})
	}
}

func TestK8sImpl_CreateK8sJobWithUserDefinedLabels(t *testing.T) {
	k8sClientSet := k8sfake.NewSimpleClientset()
	k8s := K8sImpl{clientset: k8sClientSet}

	userDefinedLabels := map[string]string{
		"TestLabel1":               "SomeKey_0",
		"some.dns.name/identifier": "value",
	}

	eventData := keptnv2.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
	}

	var event map[string]interface{}
	err := json.Unmarshal([]byte(testTriggeredEvent), &event)
	require.NoError(t, err)

	err = k8s.CreateK8sJob(
		"job-1-2-3-1",
		JobDetails{
			Action: &config.Action{
				Name: "Test Action",
			},
			Task: &config.Task{
				Name:  "Test Job",
				Image: "alpine",
				Cmd:   []string{"ls"},
			},
			ActionIndex:   0,
			TaskIndex:     0,
			JobConfigHash: "",
		},
		&eventData,
		JobSettings{
			JobNamespace: testNamespace,
			DefaultResourceRequirements: &corev1.ResourceRequirements{
				Limits:   make(corev1.ResourceList),
				Requests: make(corev1.ResourceList),
			},
			DefaultPodSecurityContext: new(corev1.PodSecurityContext),
			DefaultSecurityContext:    new(corev1.SecurityContext),
			JobLabels:                 userDefinedLabels,
		},
		event,
		testNamespace,
	)
	require.NoError(t, err)

	job, err := k8sClientSet.BatchV1().Jobs(testNamespace).Get(context.TODO(), "job-1-2-3-1", metav1.GetOptions{})
	require.NoError(t, err)

	expectedLabels := map[string]string{
		"app.kubernetes.io/managed-by": "job-executor-service",
		"keptn.sh/context":             event["shkeptncontext"].(string),
		"keptn.sh/ceid":                event["id"].(string),
		"keptn.sh/commitid":            "",
		"keptn.sh/jes-action":          "Test_Action",
		"keptn.sh/jes-task":            "Test_Job",
		"keptn.sh/job-confighash":      "",
		"keptn.sh/jes-action-index":    "0",
		"keptn.sh/jes-task-index":      "0",
	}
	for key, value := range userDefinedLabels {
		expectedLabels[key] = value
	}

	assert.Equal(t, expectedLabels, job.Labels)
}
