package keptn

import (
	"github.com/golang/mock/gomock"
	"github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils"
	"github.com/stretchr/testify/require"
	keptnfake "keptn-contrib/job-executor-service/pkg/keptn/fake"
	"testing"
)

func TestV1ResourceHandler_GetResource(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockResourceHandler := keptnfake.NewMockV1KeptnResourceHandler(mockCtrl)

	handler := V1ResourceHandler{
		Event: EventProperties{
			Project: "project",
			Stage:   "stage",
			Service: "service",
		},
		ResourceHandler: mockResourceHandler,
	}

	tests := []struct {
		Test        string
		GitCommitID string
	}{
		{
			Test:        "With GitCommitID",
			GitCommitID: "324723984372948",
		},
		{
			Test:        "Without GitCommitID",
			GitCommitID: "",
		},
	}

	for _, test := range tests {
		t.Run("GetServiceResource_"+test.Test, func(t *testing.T) {
			expectedBytes := []byte("<expected-file-payload>")

			scope := api.NewResourceScope()
			scope.Project("project")
			scope.Resource("resource")
			scope.Service("service")
			scope.Stage("stage")

			mockResourceHandler.EXPECT().GetResource(*scope, gomock.Len(1)).Times(1).Return(&models.Resource{
				Metadata:        nil,
				ResourceContent: string(expectedBytes),
				ResourceURI:     nil,
			}, nil)

			resource, err := handler.GetServiceResource("resource", test.GitCommitID)
			require.NoError(t, err)
			require.Equal(t, expectedBytes, resource)
		})

		t.Run("GetStageResource_"+test.Test, func(t *testing.T) {
			expectedBytes := []byte("<expected-file-payload>")

			scope := api.NewResourceScope()
			scope.Project("project")
			scope.Resource("resource")
			scope.Stage("stage")

			mockResourceHandler.EXPECT().GetResource(*scope, gomock.Len(1)).Times(1).Return(&models.Resource{
				Metadata:        nil,
				ResourceContent: string(expectedBytes),
				ResourceURI:     nil,
			}, nil)

			resource, err := handler.GetStageResource("resource", test.GitCommitID)
			require.NoError(t, err)
			require.Equal(t, expectedBytes, resource)
		})

		t.Run("GetProjectResource_"+test.Test, func(t *testing.T) {
			expectedBytes := []byte("<expected-file-payload>")

			scope := api.NewResourceScope()
			scope.Project("project")
			scope.Resource("resource")

			mockResourceHandler.EXPECT().GetResource(*scope, gomock.Len(1)).Times(1).Return(&models.Resource{
				Metadata:        nil,
				ResourceContent: string(expectedBytes),
				ResourceURI:     nil,
			}, nil)

			resource, err := handler.GetProjectResource("resource", test.GitCommitID)
			require.NoError(t, err)
			require.Equal(t, expectedBytes, resource)
		})
	}

}
