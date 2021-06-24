package keptn

import (
	"github.com/golang/mock/gomock"
	"github.com/keptn/go-utils/pkg/api/models"
	"github.com/spf13/afero"
	"gotest.tools/assert"
	keptnfake "keptn-sandbox/job-executor-service/pkg/keptn/fake"
	"net/url"
	"testing"
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
		}, nil)

	resourceHandlerMock.EXPECT().GetServiceResource(project, stage, service, url.QueryEscape(locustBasic)).Times(1).Return(
		&models.Resource{
			Metadata:        nil,
			ResourceContent: locustBasic,
			ResourceURI:     nil,
		}, nil)

	resourceHandlerMock.EXPECT().GetServiceResource(project, stage, service, url.QueryEscape(locustFunctional)).Times(1).Return(
		&models.Resource{
			Metadata:        nil,
			ResourceContent: locustFunctional,
			ResourceURI:     nil,
		}, nil)

	configService := NewConfigService(false, project, stage, service, resourceHandlerMock)
	fs := afero.NewMemMapFs()

	keptnResources, err := configService.GetAllKeptnResources(fs, "locust")
	assert.NilError(t, err)

	val, ok := keptnResources[locustBasic]
	assert.Assert(t, ok)
	assert.Equal(t, string(val), locustBasic)

	val, ok = keptnResources[locustFunctional]
	assert.Assert(t, ok)
	assert.Equal(t, string(val), locustFunctional)
}
