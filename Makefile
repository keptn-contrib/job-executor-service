.PHONY: build-lint build-docker-service build-docker-initcontainer build-docker

TAG ?= latest

build-lint:
	@echo "Compiling job executor service lint for every OS and Platform"
	GOOS=linux GOARCH=amd64 go build -o bin/job-lint-linux-amd64 ./cmd/job-executor-service-lint
	GOOS=windows GOARCH=amd64 go build -o bin/job-lint-windows-amd64 ./cmd/job-executor-service-lint
	GOOS=darwin GOARCH=amd64 go build -o bin/job-lint-darwin-amd64 ./cmd/job-executor-service-lint

build-docker-service:
	@echo "Building docker image keptnsandbox/job-executor-service:$(TAG)"
	docker build . -f Dockerfile -t keptnsandbox/job-executor-service:$(TAG)

build-docker-initcontainer:
	@echo "Building docker image keptnsandbox/job-executor-service-initcontainer:$(TAG)"
	docker build . -f initcontainer.Dockerfile -t keptnsandbox/job-executor-service-initcontainer:$(TAG)

build-docker: build-docker-service build-docker-initcontainer