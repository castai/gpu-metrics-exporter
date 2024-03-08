.PHONY: lint
lint: check-lint-dependencies
	golangci-lint run ./...

.PHONY: fix-lint-lint
fix-lint: check-lint-dependencies
	golangci-lint run ./... --fix

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/gpu-metrics-exporter-amd64 ./cmd/main.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/gpu-metrics-exporter-arm64 ./cmd/main.go
	docker buildx build --push --platform=linux/amd64,linux/arm64 -t  $(TAG) .

.PHONY: generate
generate:
	go generate ./...

SHELL := /bin/bash
.PHONY: run
run:
	go run ./cmd/main.go

.PHONY: test
test:
	go test ./... -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: gen-proto
gen-proto: check-proto-dependencies
	protoc pb/metrics.proto --go_out=paths=source_relative:.

.PHONY: check-lint-dependencies
check-lint-dependencies:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, please install it" && exit 1)

.PHONY: check-proto-dependencies
check-proto-dependencies:
	@which protoc > /dev/null || (echo "protoc not found, please install it" && exit 1)