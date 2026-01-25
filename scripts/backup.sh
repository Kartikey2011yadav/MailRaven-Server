#!/bin/bash
set -e

# scripts/backup.sh
# Trigger a hot backup of the MailRaven server.
# Usage: ./backup.sh <target_dir>

TARGET_DIR=${1:-/tmp/mailraven-backup}
API_URL="http://localhost:8080/api/v1" # Adjust port/path as needed
ADMIN_TOKEN="${ADMIN_TOKEN:-}"

if [ -z "$ADMIN_TOKEN" ]; then
  echo "Error: ADMIN_TOKEN environment variable is not set."
  exit 1
fi

echo "Triggering backup to: $TARGET_DIR"

# Call the Admin API (TBD implementation in T026)
curl -X POST "$API_URL/admin/backup" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"location\": \"$TARGET_DIR\"}"

echo "Backup job initiated."
