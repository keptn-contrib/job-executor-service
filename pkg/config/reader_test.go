package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"keptn-contrib/job-executor-service/pkg/config/fake"
)

func TestConfigRetrievalFailed(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockKeptnResourceService := fake.NewMockKeptnResourceService(mockCtrl)
	retrievalError := errors.New("error getting resource")
	mockKeptnResourceService.EXPECT().GetServiceResource("job/config.yaml", "c25692cb4fe4068fbdc2").Return(nil, retrievalError)
	mockKeptnResourceService.EXPECT().GetStageResource("job/config.yaml", "c25692cb4fe4068fbdc2").Return(nil, retrievalError)

	// NOTE: fetching project resources works only without a git commit id, because of branches !
	mockKeptnResourceService.EXPECT().GetProjectResource("job/config.yaml").Return(nil, retrievalError)

	sut := JobConfigReader{Keptn: mockKeptnResourceService}

	config, _, err := sut.GetJobConfig("c25692cb4fe4068fbdc2")
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestMalformedConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockKeptnResourceService := fake.NewMockKeptnResourceService(mockCtrl)
	yamlConfig := `
                    someyaml_that:
                            has_nothing_to_do:
                                with_job_executor: true
                    `
	mockKeptnResourceService.EXPECT().GetServiceResource("job/config.yaml", "").Return(
		[]byte(yamlConfig),
		nil,
	)

	sut := JobConfigReader{Keptn: mockKeptnResourceService}

	config, _, err := sut.GetJobConfig("")
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestGetConfigHappyPath(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockKeptnResourceService := fake.NewMockKeptnResourceService(mockCtrl)
	yamlConfig := `
                    apiVersion: v2
                    actions:
                      - name: "Run whatever you like with JES"
                        events:
                        tasks:
                          - name: "task1"
                            workingDir: "/bin"
                            image: "somefancyimage"
                            cmd:
                                - echo "Hello World!"
                    `
	mockKeptnResourceService.EXPECT().GetServiceResource("job/config.yaml", "c25692cb4fe4068fbdc2").Return(
		[]byte(yamlConfig),
		nil,
	)

	sut := JobConfigReader{Keptn: mockKeptnResourceService}

	config, _, err := sut.GetJobConfig("c25692cb4fe4068fbdc2")
	assert.NoError(t, err)
	assert.NotNil(t, config)
}

func TestJobConfigReader_FindJobConfigResource(t *testing.T) {

	t.Run("Find in service", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockKeptnResourceService := fake.NewMockKeptnResourceService(mockCtrl)

		sut := JobConfigReader{Keptn: mockKeptnResourceService}

		mockKeptnResourceService.EXPECT().GetServiceResource("job/config.yaml", "c25692cb4fe4068fbdc2").Return(
			[]byte("test"),
			nil,
		)

		result, err := sut.FindJobConfigResource("c25692cb4fe4068fbdc2")
		assert.NoError(t, err)
		assert.Equal(t, result, []byte("test"))
	})

	t.Run("Find in stage", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockKeptnResourceService := fake.NewMockKeptnResourceService(mockCtrl)

		sut := JobConfigReader{Keptn: mockKeptnResourceService}

		mockKeptnResourceService.EXPECT().GetServiceResource("job/config.yaml", "c25692cb4fe4068fbdc2").Return(
			nil,
			fmt.Errorf("some error"),
		)

		mockKeptnResourceService.EXPECT().GetStageResource("job/config.yaml", "c25692cb4fe4068fbdc2").Return(
			[]byte("test1"),
			nil,
		)

		result, err := sut.FindJobConfigResource("c25692cb4fe4068fbdc2")
		assert.NoError(t, err)
		assert.Equal(t, result, []byte("test1"))
	})

	t.Run("Find in project", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockKeptnResourceService := fake.NewMockKeptnResourceService(mockCtrl)

		sut := JobConfigReader{Keptn: mockKeptnResourceService}

		mockKeptnResourceService.EXPECT().GetServiceResource("job/config.yaml", "c25692cb4fe4068fbdc2").Return(
			nil,
			fmt.Errorf("some error"),
		)

		mockKeptnResourceService.EXPECT().GetStageResource("job/config.yaml", "c25692cb4fe4068fbdc2").Return(
			nil,
			fmt.Errorf("some error"),
		)

		mockKeptnResourceService.EXPECT().GetProjectResource("job/config.yaml").Return(
			[]byte("abc"),
			nil,
		)

		result, err := sut.FindJobConfigResource("c25692cb4fe4068fbdc2")
		assert.NoError(t, err)
		assert.Equal(t, result, []byte("abc"))
	})

	t.Run("Not job config", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockKeptnResourceService := fake.NewMockKeptnResourceService(mockCtrl)

		sut := JobConfigReader{Keptn: mockKeptnResourceService}

		mockKeptnResourceService.EXPECT().GetServiceResource("job/config.yaml", "c25692cb4fe4068fbdc2").Return(
			nil,
			fmt.Errorf("some error"),
		)

		mockKeptnResourceService.EXPECT().GetStageResource("job/config.yaml", "c25692cb4fe4068fbdc2").Return(
			nil,
			fmt.Errorf("some error"),
		)

		mockKeptnResourceService.EXPECT().GetProjectResource("job/config.yaml").Return(
			nil,
			fmt.Errorf("some error"),
		)

		result, err := sut.FindJobConfigResource("c25692cb4fe4068fbdc2")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

}
