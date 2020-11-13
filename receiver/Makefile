SHELL:=/usr/bin/env bash

common:
	go version
	go clean
	go mod tidy

build: common
	go build *.go

test: common
	go test -v -cover

lint:
	docker run --rm -v ${PWD}:/app -w /app golangci/golangci-lint:v1.32.0 golangci-lint run

release:
	@ chmod +x ../ci/release.sh
	@ ../ci/release.sh "package" "receiver/v1.0.2"

verify:
	go mod verify