package file

import (
	"testing"

	keptnfake "keptn-contrib/job-executor-service/pkg/keptn/fake"

	"github.com/golang/mock/gomock"
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

func CreateKeptnResourceServiceMock(t *testing.T) *keptnfake.MockResourceHandler {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	return keptnfake.NewMockResourceHandler(mockCtrl)
}

/*func TestMountFiles(t *testing.T) {

	fs := afero.NewMemMapFs()
	resourceServiceMock := CreateKeptnResourceServiceMock(t)

	config, _ := config2.NewConfig([]byte(simpleConfig))

	sut := config2.JobConfigReader{Keptn: resourceServiceMock}

	jcr, _, err := sut.GetJobConfig("")

	resourceServiceMock.EXPECT().GetJobConfiguration().Times(1).Return(config, nil)
	resourceServiceMock.EXPECT().GetAllKeptnResources(
		fs, "locust",
	).Times(1).Return(
		map[string][]byte{
			"locust/basic.py": []byte(pythonFile), "locust/functional.py": []byte(pythonFile),
		}, nil,
	)
	resourceServiceMock.EXPECT().GetAllKeptnResources(
		fs, "/helm/values.yaml",
	).Times(1).Return(map[string][]byte{"helm/values.yaml": []byte(yamlFile)}, nil)

	err := MountFiles("action", "task", fs, resourceServiceMock)
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
	configServiceMock := CreateKeptnResourceServiceMock(t)

	configServiceMock.EXPECT().GetJobConfiguration().Times(1).Return(nil, errors.New("not found"))

	err := MountFiles("action", "task", fs, configServiceMock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMountFilesConfigFileNotValid(t *testing.T) {

	fs := afero.NewMemMapFs()
	configServiceMock := CreateKeptnResourceServiceMock(t)

	config, configErr := config2.NewConfig([]byte(pythonFile))
	configServiceMock.EXPECT().GetJobConfiguration().Times(1).Return(config, configErr)

	err := MountFiles("action", "task", fs, configServiceMock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot unmarshal")
}

func TestMountFilesNoActionMatch(t *testing.T) {

	fs := afero.NewMemMapFs()
	configServiceMock := CreateKeptnResourceServiceMock(t)

	config, _ := config2.NewConfig([]byte(simpleConfig))
	configServiceMock.EXPECT().GetJobConfiguration().Times(1).Return(config, nil)

	err := MountFiles("actionNotMatching", "task", fs, configServiceMock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no action found with name 'actionNotMatching'")
}

func TestMountFilesNoTaskMatch(t *testing.T) {

	fs := afero.NewMemMapFs()
	configServiceMock := CreateKeptnResourceServiceMock(t)

	config, _ := config2.NewConfig([]byte(simpleConfig))
	configServiceMock.EXPECT().GetJobConfiguration().Times(1).Return(config, nil)

	err := MountFiles("action", "taskNotMatching", fs, configServiceMock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no task found with name 'taskNotMatching'")
}

func TestMountFilesFileNotFound(t *testing.T) {

	fs := afero.NewMemMapFs()
	configServiceMock := CreateKeptnResourceServiceMock(t)

	config, _ := config2.NewConfig([]byte(simpleConfig))
	configServiceMock.EXPECT().GetJobConfiguration().Times(1).Return(config, nil)
	configServiceMock.EXPECT().GetAllKeptnResources(fs, "/helm/values.yaml").Times(1).Return(
		nil, errors.New("not found"),
	)

	err := MountFiles("action", "task", fs, configServiceMock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}*/
