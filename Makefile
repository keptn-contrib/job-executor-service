.PHONY: build-lint build-docker-service build-docker-initcontainer build-docker

TAG ?= latest

generate-mocks:
	@echo "(Re)Generating mocks in */fake packages"
	go generate ./...

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