.PHONY: build run test clean docker-build docker-push helm-lint helm-package

# Variables
APP_NAME=helm-s3-exporter
VERSION?=0.1.0
DOCKER_REGISTRY?=docker.io
DOCKER_IMAGE=$(DOCKER_REGISTRY)/$(APP_NAME)
DOCKER_TAG=$(VERSION)

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	go build -ldflags="-w -s" -o bin/exporter ./cmd/exporter

# Run the application locally
run:
	@echo "Running $(APP_NAME)..."
	go run ./cmd/exporter

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out

# Build Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

# Push Docker image
docker-push:
	@echo "Pushing Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

# Lint Helm chart
helm-lint:
	@echo "Linting Helm chart..."
	helm lint charts/$(APP_NAME)

# Package Helm chart
helm-package:
	@echo "Packaging Helm chart..."
	helm package charts/$(APP_NAME) -d charts/

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -w .
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Generate mocks (if needed in future)
generate:
	@echo "Generating code..."
	go generate ./...

