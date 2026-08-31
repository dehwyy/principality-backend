#!/usr/bin/env bash
set -euo pipefail

MEDIA_DIR="${1:-../../media/principalities}"
MINIO_NAMESPACE="${MINIO_NAMESPACE:-aioffice}"
MINIO_ENDPOINT="${MINIO_ENDPOINT:-http://minio.aioffice.svc.cluster.local:9000}"
MINIO_SECRET_NAME="${MINIO_SECRET_NAME:-minio-credentials}"
BUCKET="principality-media"
UPLOADER_POD="principality-media-uploader"

kubectl -n "$MINIO_NAMESPACE" delete pod "$UPLOADER_POD" --ignore-not-found --wait=true

kubectl -n "$MINIO_NAMESPACE" run "$UPLOADER_POD" --image=minio/mc:latest --restart=Never \
  --overrides="{\"spec\":{\"containers\":[{\"name\":\"$UPLOADER_POD\",\"image\":\"minio/mc:latest\",\"command\":[\"sleep\",\"900\"],\"env\":[{\"name\":\"MINIO_ACCESS\",\"valueFrom\":{\"secretKeyRef\":{\"name\":\"$MINIO_SECRET_NAME\",\"key\":\"accessKey\"}}},{\"name\":\"MINIO_SECRET\",\"valueFrom\":{\"secretKeyRef\":{\"name\":\"$MINIO_SECRET_NAME\",\"key\":\"secretKey\"}}}]}]}}"

kubectl -n "$MINIO_NAMESPACE" wait --for=condition=Ready "pod/$UPLOADER_POD" --timeout=120s

kubectl -n "$MINIO_NAMESPACE" exec "$UPLOADER_POD" -- sh -c \
  "mc alias set media $MINIO_ENDPOINT \"\$MINIO_ACCESS\" \"\$MINIO_SECRET\" >/dev/null && \
   mc mb --ignore-existing media/$BUCKET && \
   mc anonymous set download media/$BUCKET"

for media_file in "$MEDIA_DIR"/*; do
  media_key="$(basename "$media_file")"
  case "$media_key" in
    *.jpg) content_type=image/jpeg ;;
    *.mp4) content_type=video/mp4 ;;
    *) content_type=application/octet-stream ;;
  esac
  kubectl -n "$MINIO_NAMESPACE" exec -i "$UPLOADER_POD" -- \
    mc pipe --attr "Content-Type=$content_type" "media/$BUCKET/$media_key" < "$media_file" >/dev/null
  echo "залито $media_key ($content_type)"
done

kubectl -n "$MINIO_NAMESPACE" exec "$UPLOADER_POD" -- mc ls "media/$BUCKET"
kubectl -n "$MINIO_NAMESPACE" delete pod "$UPLOADER_POD" --wait=false
