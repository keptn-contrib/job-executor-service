package keptn

import (
	api "github.com/keptn/go-utils/pkg/api/utils/v2"
	"net/url"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	keptnfake "keptn-contrib/job-executor-service/pkg/keptn/fake"

	"github.com/golang/mock/gomock"
	"github.com/keptn/go-utils/pkg/api/models"
	"github.com/spf13/afero"
)

func CreateResourceHandlerMock(t *testing.T) *keptnfake.MockV2ResourceHandler {
	mockCtrl := gomock.NewController(t)
	return keptnfake.NewMockV2ResourceHandler(mockCtrl)
}

const service = "carts"
const project = "sockshop"
const stage = "dev"
const gitCommitID = "6caf78d2c978f7f787"

func TestGetAllKeptnResources(t *testing.T) {
	locustBasic := "locust/basic.py"
	locustFunctional := "locust/functional.py"

	resourceHandlerMock := CreateResourceHandlerMock(t)
	resourceHandlerMock.EXPECT().GetAllServiceResources(gomock.Any(), project, stage, service, gomock.Any()).Times(1).Return(
		[]*models.Resource{
			{
				Metadata:        nil,
				ResourceContent: "",
				ResourceURI:     &locustBasic,
			},
			{
				Metadata:        nil,
				ResourceContent: "",
				ResourceURI:     &locustFunctional,
			},
		}, nil,
	)

	scope1 := api.NewResourceScope()
	scope1.Project(project)
	scope1.Stage(stage)
	scope1.Service(service)
	scope1.Resource(locustBasic)

	resourceHandlerMock.EXPECT().GetResource(gomock.Any(), *scope1, gomock.Any()).Times(1).Return(
		&models.Resource{
			Metadata:        nil,
			ResourceContent: locustBasic,
			ResourceURI:     nil,
		}, nil,
	)

	scope2 := api.NewResourceScope()
	scope2.Project(project)
	scope2.Stage(stage)
	scope2.Service(service)
	scope2.Resource(locustFunctional)

	resourceHandlerMock.EXPECT().GetResource(gomock.Any(), *scope2, gomock.Any()).Times(1).Return(
		&models.Resource{
			Metadata:        nil,
			ResourceContent: locustFunctional,
			ResourceURI:     nil,
		}, nil,
	)

	event := EventProperties{
		Project:     project,
		Stage:       stage,
		Service:     service,
		GitCommitID: gitCommitID,
	}

	configService := NewConfigService(false, event, resourceHandlerMock)
	fs := afero.NewMemMapFs()

	keptnResources, err := configService.GetAllKeptnResources(fs, "locust")
	require.NoError(t, err)

	val, ok := keptnResources[locustBasic]
	assert.True(t, ok)
	assert.Equal(t, string(val), locustBasic)

	val, ok = keptnResources[locustFunctional]
	assert.True(t, ok)
	assert.Equal(t, string(val), locustFunctional)
}

func TestGetAllKeptnResourcesLocal(t *testing.T) {
	locustPath := "/locust"
	locustBasic := path.Join(locustPath, "basic.py")
	locustFunctional := path.Join(locustPath, "functional.py")

	resourceHandlerMock := CreateResourceHandlerMock(t)
	resourceHandlerMock.EXPECT().GetAllServiceResources(gomock.Any(), project, stage, service, gomock.Any()).Times(0)

	scope1 := api.NewResourceScope()
	scope1.Project(project)
	scope1.Stage(stage)
	scope1.Service(service)
	scope1.Resource(url.QueryEscape(locustBasic))

	resourceHandlerMock.EXPECT().GetResource(gomock.Any(), scope1, gomock.Any()).Times(0)

	scope2 := api.NewResourceScope()
	scope2.Project(project)
	scope2.Stage(stage)
	scope2.Service(service)
	scope2.Resource(url.QueryEscape(locustFunctional))

	resourceHandlerMock.EXPECT().GetResource(gomock.Any(), scope2, gomock.Any()).Times(0)

	event := EventProperties{
		Project:     project,
		Stage:       stage,
		Service:     service,
		GitCommitID: "",
	}

	configService := NewConfigService(true, event, resourceHandlerMock)
	fs := afero.NewMemMapFs()

	err := createFile(fs, locustBasic, []byte(locustBasic))
	require.NoError(t, err)

	err = createFile(fs, locustFunctional, []byte(locustFunctional))
	require.NoError(t, err)

	keptnResources, err := configService.GetAllKeptnResources(fs, locustPath)
	require.NoError(t, err)

	val, ok := keptnResources[locustBasic]
	assert.True(t, ok)
	assert.Equal(t, locustBasic, string(val))

	val, ok = keptnResources[locustFunctional]
	assert.True(t, ok)
	assert.Equal(t, string(val), locustFunctional)
}

func TestErrorNoDirectoryResourcesLocal(t *testing.T) {
	locustPath := "/locust"
	locustBasic := path.Join(locustPath, "basic.py")
	locustFunctional := path.Join(locustPath, "functional.py")

	resourceHandlerMock := CreateResourceHandlerMock(t)
	resourceHandlerMock.EXPECT().GetAllServiceResources(gomock.Any(), project, stage, service, gomock.Any()).Times(0)

	scope1 := api.NewResourceScope()
	scope1.Project(project)
	scope1.Stage(stage)
	scope1.Service(service)
	scope1.Resource(url.QueryEscape(locustBasic))

	resourceHandlerMock.EXPECT().GetResource(gomock.Any(), scope1, gomock.Any()).Times(0)

	scope2 := api.NewResourceScope()
	scope2.Project(project)
	scope2.Stage(stage)
	scope2.Service(service)
	scope2.Resource(url.QueryEscape(locustFunctional))

	resourceHandlerMock.EXPECT().GetResource(gomock.Any(), scope2, gomock.Any()).Times(0)

	event := EventProperties{
		Project:     project,
		Stage:       stage,
		Service:     service,
		GitCommitID: "",
	}

	configService := NewConfigService(true, event, resourceHandlerMock)
	fs := afero.NewMemMapFs()

	_, err := configService.GetAllKeptnResources(fs, locustPath)
	require.Error(t, err)
}

func createFile(fs afero.Fs, fileName string, content []byte) error {
	file, err := fs.Create(fileName)
	if err != nil {
		return err
	}
	file.Write(content)
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}
