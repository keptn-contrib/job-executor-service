package keptn

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils/v2"
	"github.com/stretchr/testify/require"
	keptnfake "keptn-contrib/job-executor-service/pkg/keptn/fake"
	"testing"
)

func TestV1ResourceHandler_GetResource(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockResourceAPI := keptnfake.NewMockResourcesInterface(mockCtrl)

	handler := V1ResourceHandler{
		Event: EventProperties{
			Project: "project",
			Stage:   "stage",
			Service: "service",
		},
		ResourceAPI: mockResourceAPI,
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

			mockResourceAPI.EXPECT().GetResource(context.Background(), *scope, gomock.Any()).Times(1).Return(&models.Resource{
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

			mockResourceAPI.EXPECT().GetResource(context.Background(), *scope, gomock.Any()).Times(1).Return(&models.Resource{
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

			mockResourceAPI.EXPECT().GetResource(context.Background(), *scope, gomock.Any()).Times(1).Return(&models.Resource{
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
