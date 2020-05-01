SHELL:=/bin/bash

release:
	version="$(go run version.go)"
	hub release create -m "Azure Event Hubs Lib $(version)" -m "$(git --no-pager log --oneline --decorate --color -n 5 v1.0.0..v1.0.3)" "$(version)"