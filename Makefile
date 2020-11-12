SHELL:=/bin/bash

release:
	@ chmod +x ./ci/release.sh
	@ ./ci/release.sh ${PWD}/version.go

build:
	go build -v .