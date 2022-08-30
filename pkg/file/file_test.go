package file

import (
	"errors"
	config2 "keptn-contrib/job-executor-service/pkg/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	configfake "keptn-contrib/job-executor-service/pkg/config/fake"
)

const simpleConfig = `
apiVersion: v2
actions:
  - name: "action"
    events:
      - name: "sh.keptn.event.test.triggered"
        jsonpath:
          property: "$.test.teststrategy" 
          match: "health"
    tasks:
      - name: "task"
        files:
          - /helm/values.yaml
          - locust
        image: "locustio/locust"
        cmd:
          - locust
        args:
          - '-f'
          - /keptn/locust/basic.py
`

const pythonFile = `
// This is a python file
`

const yamlFile = `
// This is a yaml file
`

func CreateKeptnResourceServiceMock(t *testing.T) *configfake.MockKeptnResourceService {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	return configfake.NewMockKeptnResourceService(mockCtrl)
}

func TestMountFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	resourceServiceMock := CreateKeptnResourceServiceMock(t)

	resourceServiceMock.EXPECT().GetServiceResource("job/config.yaml", "").Return(
		[]byte(simpleConfig),
		nil,
	)

	sut := config2.JobConfigReader{Keptn: resourceServiceMock}

	resourceServiceMock.EXPECT().GetAllKeptnResources(
		"locust",
	).Times(1).Return(
		map[string][]byte{
			"locust/basic.py": []byte(pythonFile), "locust/functional.py": []byte(pythonFile),
		}, nil,
	)
	resourceServiceMock.EXPECT().GetAllKeptnResources(
		"/helm/values.yaml",
	).Times(1).Return(map[string][]byte{"helm/values.yaml": []byte(yamlFile)}, nil)

	err := MountFiles("action", "task", "", fs, sut)
	require.NoError(t, err)

	exists, err := afero.Exists(fs, "/keptn/locust/basic.py")
	assert.NoError(t, err)
	assert.True(t, exists)

	file, err := afero.ReadFile(fs, "/keptn/locust/basic.py")
	assert.NoError(t, err)
	assert.Equal(t, pythonFile, string(file))

	exists, err = afero.Exists(fs, "/keptn/helm/values.yaml")
	assert.NoError(t, err)
	assert.True(t, exists)

	file, err = afero.ReadFile(fs, "/keptn/helm/values.yaml")
	assert.NoError(t, err)
	assert.Equal(t, yamlFile, string(file))
}

func TestMountFilesConfigFileNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()
	resourceServiceMock := CreateKeptnResourceServiceMock(t)

	resourceServiceMock.EXPECT().GetServiceResource("job/config.yaml", "").Return(
		nil,
		nil,
	)

	sut := config2.JobConfigReader{Keptn: resourceServiceMock}

	err := MountFiles("action", "task", "", fs, sut)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not find config for job-executor-service")
}

func TestMountFilesConfigFileNotValid(t *testing.T) {
	fs := afero.NewMemMapFs()
	resourceServiceMock := CreateKeptnResourceServiceMock(t)

	resourceServiceMock.EXPECT().GetServiceResource("job/config.yaml", "").Return(
		[]byte(pythonFile),
		nil,
	)

	sut := config2.JobConfigReader{Keptn: resourceServiceMock}

	err := MountFiles("action", "task", "", fs, sut)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot unmarshal")
}

func TestMountFilesNoActionMatch(t *testing.T) {
	fs := afero.NewMemMapFs()
	resourceServiceMock := CreateKeptnResourceServiceMock(t)

	resourceServiceMock.EXPECT().GetServiceResource("job/config.yaml", "").Return(
		[]byte(simpleConfig),
		nil,
	)

	sut := config2.JobConfigReader{Keptn: resourceServiceMock}

	err := MountFiles("actionNotMatching", "task", "", fs, sut)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no action found with name 'actionNotMatching'")
}

func TestMountFilesNoTaskMatch(t *testing.T) {
	fs := afero.NewMemMapFs()
	resourceServiceMock := CreateKeptnResourceServiceMock(t)

	resourceServiceMock.EXPECT().GetServiceResource("job/config.yaml", "").Return(
		[]byte(simpleConfig),
		nil,
	)

	sut := config2.JobConfigReader{Keptn: resourceServiceMock}

	err := MountFiles("action", "taskNotMatching", "", fs, sut)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no task found with name 'taskNotMatching'")
}

func TestMountFilesFileNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()
	resourceServiceMock := CreateKeptnResourceServiceMock(t)

	resourceServiceMock.EXPECT().GetServiceResource("job/config.yaml", "").Return(
		[]byte(simpleConfig),
		nil,
	)

	resourceServiceMock.EXPECT().GetAllKeptnResources(
		"/helm/values.yaml",
	).Times(1).Return(nil, errors.New("not found"))

	sut := config2.JobConfigReader{Keptn: resourceServiceMock}

	err := MountFiles("action", "task", "", fs, sut)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
