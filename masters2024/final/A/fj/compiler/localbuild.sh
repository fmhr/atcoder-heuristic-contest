#!/bin/sh

IMAGE_NAME="local-compiler"

set -ex
docker build -t ${IMAGE_NAME} .
docker run -p 8080:8080 ${IMAGE_NAME}