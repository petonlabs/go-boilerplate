#!/bin/sh
set -euo pipefail

TS="$(date -u +%Y%m%dT%H%M%SZ)"
DIR="/var/backups"
FILE="${DIR}/pg_${DATABASE_NAME}_${TS}.sql.zst"

mkdir -p "$DIR"

echo "[backup] Dumping database ${DATABASE_NAME} at ${TS}"
pg_dump -h "${DATABASE_HOST}" -p "${DATABASE_PORT}" -U "${DATABASE_USER}" -d "${DATABASE_NAME}" --no-owner --no-privileges | zstd -T0 -o "$FILE"

# Upload to Cloudflare R2 (S3-compatible)
echo "[backup] Uploading to S3 ${S3_BUCKET} via ${S3_ENDPOINT}"
aws --endpoint-url "${S3_ENDPOINT}" s3 cp "$FILE" "s3://${S3_BUCKET}/postgres/${DATABASE_NAME}/$(basename "$FILE")" --no-progress

# Retention: delete older than N days locally
find "$DIR" -type f -name "pg_${DATABASE_NAME}_*.sql.zst" -mtime +${BACKUP_RETENTION_DAYS:-14} -delete || true
echo "[backup] Done"
