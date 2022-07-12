#!/bin/bash

# Fetch latest release and prerelease
prerelease=$(curl -s https://api.github.com/repos/keptn/keptn/releases | jq -r 'map(select(.prerelease)) | first | .tag_name')
release=$(curl -s https://api.github.com/repos/keptn/keptn/releases | jq -r 'map(select(.prerelease==false)) | first | .tag_name')

matrix_config=""

# Write variables as output
echo "::set-output name=LATEST_RELEASE::$release"
echo "::set-output name=LATEST_PRERELEASE::$prerelease"