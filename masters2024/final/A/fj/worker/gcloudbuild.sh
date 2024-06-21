#!/bin/sh

GCLOUD_PROJECT="ahc-contests"
REPOSITORY="my-app-images"
IMAGE_NAME="ahc023-tester"
REGION="asia-northeast1"

set -e #コマンドが失敗したらそこで終了する
set -x #実行したコマンドを表示する

gcloud builds submit --tag ${REGION}-docker.pkg.dev/${GCLOUD_PROJECT}/${REPOSITORY}/${IMAGE_NAME}:latest .

gcloud run deploy ${IMAGE_NAME}\
    --image ${REGION}-docker.pkg.dev/${GCLOUD_PROJECT}/${REPOSITORY}/${IMAGE_NAME}:latest \
    --platform managed --allow-unauthenticated \
    --concurrency 1