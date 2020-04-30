SHELL:=/usr/bin/env bash

common:
	go version
	go clean

build: common
	go build *.go

test: common
	go test -v -cover

lint:
	docker run --rm -v ${PWD}:/app -w /app golangci/golangci-lint:v1.25.1 golangci-lint run

verify:
	go mod verify