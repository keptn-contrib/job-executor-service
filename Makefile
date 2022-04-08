.PHONY: build-lint build-docker-service build-docker-initcontainer build-docker

TAG ?= latest

generate-mocks:
	@echo "(Re)Generating mocks in */fake packages"
	mockgen -source=pkg/k8sutils/connect.go -destination=pkg/k8sutils/fake/connect_mock.go -package fake
	mockgen -source=pkg/keptn/config_service.go -destination=pkg/keptn/fake/config_service_mock.go -package fake
	mockgen -source=pkg/eventhandler/eventhandlers.go -destination=pkg/eventhandler/fake/eventhandlers_mock.go -package fake

build-lint:
	@echo "Compiling job executor service lint for every OS and Platform"
	GOOS=linux GOARCH=amd64 go build -o bin/job-lint-linux-amd64 ./cmd/job-executor-service-lint
	GOOS=windows GOARCH=amd64 go build -o bin/job-lint-windows-amd64 ./cmd/job-executor-service-lint
	GOOS=darwin GOARCH=amd64 go build -o bin/job-lint-darwin-amd64 ./cmd/job-executor-service-lint

build-docker-service:
	@echo "Building docker image keptncontrib/job-executor-service:$(TAG)"
	docker build . -f Dockerfile -t keptncontrib/job-executor-service:$(TAG)

build-docker-initcontainer:
	@echo "Building docker image keptncontrib/job-executor-service-initcontainer:$(TAG)"
	docker build . -f initcontainer.Dockerfile -t keptncontrib/job-executor-service-initcontainer:$(TAG)

build-docker: build-docker-service build-docker-initcontainer