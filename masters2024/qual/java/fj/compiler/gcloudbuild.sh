#!/bin/sh

GCLOUD_PROJECT="project2323"
IMAGE_NAME="fj-compiler"

set -e #コマンドが失敗したらそこで終了する
set -x #実行したコマンドを表示する

gcloud builds submit --tag gcr.io/${GCLOUD_PROJECT}/${IMAGE_NAME}:latest .
gcloud run deploy go-compiler\
    --image gcr.io/${GCLOUD_PROJECT}/${IMAGE_NAME}:latest \
    --platform managed --allow-unauthenticated