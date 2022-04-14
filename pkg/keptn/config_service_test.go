package keptn

import (
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

func CreateResourceHandlerMock(t *testing.T) *keptnfake.MockResourceHandler {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	return keptnfake.NewMockResourceHandler(mockCtrl)
}

const service = "carts"
const project = "sockshop"
const stage = "dev"

func TestGetAllKeptnResources(t *testing.T) {
	locustBasic := "/locust/basic.py"
	locustFunctional := "/locust/functional.py"

	resourceHandlerMock := CreateResourceHandlerMock(t)
	resourceHandlerMock.EXPECT().GetAllServiceResources(project, stage, service).Times(1).Return(
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

	resourceHandlerMock.EXPECT().GetServiceResource(
		project, stage, service, url.QueryEscape(locustBasic),
	).Times(1).Return(
		&models.Resource{
			Metadata:        nil,
			ResourceContent: locustBasic,
			ResourceURI:     nil,
		}, nil,
	)

	resourceHandlerMock.EXPECT().GetServiceResource(
		project, stage, service, url.QueryEscape(locustFunctional),
	).Times(1).Return(
		&models.Resource{
			Metadata:        nil,
			ResourceContent: locustFunctional,
			ResourceURI:     nil,
		}, nil,
	)

	configService := NewConfigService(false, project, stage, service, resourceHandlerMock)
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
	resourceHandlerMock.EXPECT().GetAllServiceResources(project, stage, service).Times(0)

	resourceHandlerMock.EXPECT().GetServiceResource(
		project, stage, service, url.QueryEscape(locustBasic),
	).Times(0)

	resourceHandlerMock.EXPECT().GetServiceResource(
		project, stage, service, url.QueryEscape(locustFunctional),
	).Times(0)

	configService := NewConfigService(true, project, stage, service, resourceHandlerMock)
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
	resourceHandlerMock.EXPECT().GetAllServiceResources(project, stage, service).Times(0)

	resourceHandlerMock.EXPECT().GetServiceResource(
		project, stage, service, url.QueryEscape(locustBasic),
	).Times(0)

	resourceHandlerMock.EXPECT().GetServiceResource(
		project, stage, service, url.QueryEscape(locustFunctional),
	).Times(0)

	configService := NewConfigService(true, project, stage, service, resourceHandlerMock)
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
