.PHONY: fmt lint fix build all

fmt:
	gofumpt -w .

lint:
	golangci-lint run ./...

fix:
	golangci-lint run --fix ./...

build: lint
	go build -o ugm .

all: fmt fix build
