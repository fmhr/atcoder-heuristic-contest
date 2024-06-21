#!/bin/sh

GCLOUD_PROJECT="project2323"
IMAGE_NAME="fj-worker"

set -e #コマンドが失敗したらそこで終了する
set -x #実行したコマンドを表示する

gcloud builds submit --tag gcr.io/$GCLOUD_PROJECT/$IMAGE_NAME .
gcloud run deploy worker\
    --image gcr.io/$GCLOUD_PROJECT/$IMAGE_NAME \
    --platform managed --allow-unauthenticated \
    --concurrency 1
