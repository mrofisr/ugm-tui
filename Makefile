.PHONY: fmt lint fix build

fmt:
	gofumpt -w .

lint:
	golangci-lint run ./...

fix:
	golangci-lint run --fix ./...

build:
	go build -o ugm .

all: fmt fix build
