SHELL:=/bin/bash

version = $(shell go run version.go)
latestTag = $(shell git --no-pager tag -l | tail -1)

release:
	git --no-pager log --oneline --decorate --color -n 5 "$(latestTag)"...HEAD > change-log.log
	hub release create -m "Azure Event Hubs Lib $(version)" -F change-log.log "$(version)"