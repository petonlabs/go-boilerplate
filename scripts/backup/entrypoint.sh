#!/bin/sh
set -euo pipefail

# Install needed tools
apk add --no-cache postgresql-client aws-cli tzdata ca-certificates zstd

# Write cronjob
CRON="${BACKUP_CRON:-0 */6 * * *} /bin/sh /backup/run-backup.sh >> /var/log/backup.log 2>&1"
echo "$CRON" > /etc/crontabs/root

# Start cron in foreground
touch /var/log/backup.log
crond -f -l 2
