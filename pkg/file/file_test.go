package file

import (
	"didiladi/keptn-generic-job-service/pkg/keptn"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"gotest.tools/assert"
	"testing"
)

const simpleConfig = `
actions:
  - name: "action"
    event: "sh.keptn.event.test.triggered"
    jsonpath:
      property: "$.test.teststrategy" 
      match: "health"
    tasks:
      - name: "task"
        files: 
          - locust/basic.py
        image: "locustio/locust"
        cmd: "locust -f /keptn/locust/locustfile.py"
`

const pythonFile = `
// This is a python file
`

const escapedSlash = "%2F"

func CreateKeptnConfigServiceMock(t *testing.T) *keptn.MockConfigService {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	return keptn.NewMockConfigService(mockCtrl)
}

func TestMountFiles(t *testing.T) {

	fs := afero.NewMemMapFs()
	configServiceMock := CreateKeptnConfigServiceMock(t)

	configServiceMock.EXPECT().GetKeptnResource("generic-job"+escapedSlash+"config.yaml").Times(1).Return([]byte(simpleConfig), nil)
	configServiceMock.EXPECT().GetKeptnResource("locust"+escapedSlash+"basic.py").Times(1).Return([]byte(pythonFile), nil)

	err := MountFiles("action", "task", fs, configServiceMock)
	assert.NilError(t, err)

	exists, err := afero.Exists(fs, "/keptn/locust/basic.py")
	assert.NilError(t, err)
	assert.Check(t, exists)

	file, err := afero.ReadFile(fs, "/keptn/locust/basic.py")
	assert.NilError(t, err)
	assert.Equal(t, pythonFile, string(file))
}

func TestMountFilesConfigFileNotFound(t *testing.T) {

	fs := afero.NewMemMapFs()
	configServiceMock := CreateKeptnConfigServiceMock(t)

	configServiceMock.EXPECT().GetKeptnResource("generic-job"+escapedSlash+"config.yaml").Times(1).Return(nil, errors.New("not found"))

	err := MountFiles("action", "task", fs, configServiceMock)
	assert.ErrorContains(t, err, "not found")
}

func TestMountFilesConfigFileNotValid(t *testing.T) {

	fs := afero.NewMemMapFs()
	configServiceMock := CreateKeptnConfigServiceMock(t)

	configServiceMock.EXPECT().GetKeptnResource("generic-job"+escapedSlash+"config.yaml").Times(1).Return([]byte(pythonFile), nil)

	err := MountFiles("action", "task", fs, configServiceMock)
	assert.ErrorContains(t, err, "cannot unmarshal")
}

func TestMountFilesNoActionMatch(t *testing.T) {

	fs := afero.NewMemMapFs()
	configServiceMock := CreateKeptnConfigServiceMock(t)

	configServiceMock.EXPECT().GetKeptnResource("generic-job"+escapedSlash+"config.yaml").Times(1).Return([]byte(simpleConfig), nil)

	err := MountFiles("actionNotMatching", "task", fs, configServiceMock)
	assert.ErrorContains(t, err, "no action found with name 'actionNotMatching'")
}

func TestMountFilesNoTaskMatch(t *testing.T) {

	fs := afero.NewMemMapFs()
	configServiceMock := CreateKeptnConfigServiceMock(t)

	configServiceMock.EXPECT().GetKeptnResource("generic-job"+escapedSlash+"config.yaml").Times(1).Return([]byte(simpleConfig), nil)

	err := MountFiles("action", "taskNotMatching", fs, configServiceMock)
	assert.ErrorContains(t, err, "no task found with name 'taskNotMatching'")
}

func TestMountFilesFileNotFound(t *testing.T) {

	fs := afero.NewMemMapFs()
	configServiceMock := CreateKeptnConfigServiceMock(t)

	configServiceMock.EXPECT().GetKeptnResource("generic-job"+escapedSlash+"config.yaml").Times(1).Return([]byte(simpleConfig), nil)
	configServiceMock.EXPECT().GetKeptnResource("locust"+escapedSlash+"basic.py").Times(1).Return(nil, errors.New("not found"))

	err := MountFiles("action", "task", fs, configServiceMock)
	assert.ErrorContains(t, err, "not found")
}
