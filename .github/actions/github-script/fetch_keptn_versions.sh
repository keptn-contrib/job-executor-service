#!/bin/bash

# Fetch latest release and prerelease
prerelease=$(curl -s https://api.github.com/repos/keptn/keptn/releases | jq -r 'map(select(.prerelease)) | first | .tag_name')
release=$(curl -s https://api.github.com/repos/keptn/keptn/releases | jq -r 'map(select(.prerelease==false)) | first | .tag_name')

echo $prerelease
echo $release

matrix_config=""

echo "::set-output name=LATEST_RELEASE::$release"
echo "::set-output name=LATEST_PRERELEASE::$prerelease"