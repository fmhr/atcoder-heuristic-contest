#!/bin/sh

IMAGE_NAME="fj-worker"

set -xe
docker build -t ${IMAGE_NAME} .
docker run -p 8081:8080 ${IMAGE_NAME}