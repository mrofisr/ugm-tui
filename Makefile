.PHONY: fmt lint fix build all

fmt:
	gofumpt -w .

lint:
	golangci-lint run ./...

fix:
	golangci-lint run --fix ./...

build: lint
	go build -ldflags "-X main.version=$$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o ugm ./cmd/ugm

all: fmt fix build
