#!/bin/bash

# Fetch latest release and prerelease
prerelease=$(curl -s https://api.github.com/repos/keptn/keptn/releases | jq -r 'map(select(.prerelease)) | first | .tag_name')
release=$(curl -s https://api.github.com/repos/keptn/keptn/releases | jq -r 'map(select(.prerelease==false)) | first | .tag_name')

echo $prerelease
echo $release

matrix_config=""

# Check if release or prerelease are empty
if [[ -n "$release" && -n "$prerelease" ]]
then
    # Build a matrix JSON string
    matrix_config=$(echo "{\"keptn-version\":[\"$prerelease\", \"$release\"]}")
    echo "::set-output name=KEPTN_MATRIX::$matrix_config"
else
    echo "Release or Prerelease not found"
    echo "::set-output name=KEPTN_MATRIX::{}"
fi


