#!/usr/bin/env bash

versionFile=$1

version=$(go run "$versionFile")
latestTag="$(git --no-pager tag -l | tail -1)"
changeLog="$(git --no-pager log --oneline "$latestTag" ...HEAD)"

hub release create -m "Azure Event Hubs Lib $version" -m "$changeLog" "$version"