#!/bin/sh
set -euo pipefail

if [ $# -ne 1 ]; then
  echo "Usage: restore.sh s3://bucket/path/to/pg_db_YYYYMMDDTHHMMSSZ.sql.zst"
  exit 1
fi

S3_URI="$1"
TMP="/tmp/restore.sql.zst"

echo "[restore] Downloading $S3_URI"
aws --endpoint-url "${S3_ENDPOINT}" s3 cp "$S3_URI" "$TMP" --no-progress

echo "[restore] Restoring into ${DATABASE_NAME}"
zstdcat "$TMP" | psql -h "${DATABASE_HOST}" -p "${DATABASE_PORT}" -U "${DATABASE_USER}" -d "${DATABASE_NAME}"

echo "[restore] Done"
