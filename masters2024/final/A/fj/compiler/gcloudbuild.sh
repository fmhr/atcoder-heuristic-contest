#!/bin/sh

GCLOUD_PROJECT="ahc-contests"
REPOSITORY="my-app-images"
IMAGE_NAME="compiler-go"

set -e #コマンドが失敗したらそこで終了する
set -x #実行したコマンドを表示する

gcloud builds submit --tag asia-northeast1-docker.pkg.dev/${GCLOUD_PROJECT}/${REPOSITORY}/${IMAGE_NAME}:latest .
gcloud run deploy ${IMAGE_NAME}\
    --image asia-northeast1-docker.pkg.dev/${GCLOUD_PROJECT}/${REPOSITORY}/${IMAGE_NAME}:latest \
    --platform managed
