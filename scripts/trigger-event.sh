#!/bin/bash

set -e

curl -X POST -H "Content-Type: application/cloudevents+json" localhost:8080 -d @test-events/action.triggered.json