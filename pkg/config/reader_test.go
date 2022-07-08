package config

import (
	"errors"
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
	mockKeptnResourceService.EXPECT().GetResource("job/config.yaml", "c25692cb4fe4068fbdc2").Return(nil, retrievalError)

	sut := JobConfigReader{Keptn: mockKeptnResourceService}

	config, _, err := sut.GetJobConfig("c25692cb4fe4068fbdc2")
	assert.ErrorIs(t, err, retrievalError)
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
	mockKeptnResourceService.EXPECT().GetResource("job/config.yaml", "").Return(
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
	mockKeptnResourceService.EXPECT().GetResource("job/config.yaml", "c25692cb4fe4068fbdc2").Return(
		[]byte(yamlConfig),
		nil,
	)

	sut := JobConfigReader{Keptn: mockKeptnResourceService}

	config, _, err := sut.GetJobConfig("c25692cb4fe4068fbdc2")
	assert.NoError(t, err)
	assert.NotNil(t, config)
}
