#!/bin/bash
set -e

# scripts/restore.sh
# Restore MailRaven data from a backup.
# Usage: ./restore.sh <backup_dir> <data_dir>

BACKUP_DIR=$1
DATA_DIR=$2

if [ -z "$BACKUP_DIR" ] || [ -z "$DATA_DIR" ]; then
  echo "Usage: $0 <backup_dir> <data_dir>"
  exit 1
fi

if [ ! -f "$BACKUP_DIR/meta.db" ]; then
  echo "Error: Backup directory does not contain meta.db"
  exit 1
fi

echo "Restoring from $BACKUP_DIR to $DATA_DIR..."

# Stop the server before restoring if running directly on host (omitted)

# Restore DB
cp "$BACKUP_DIR/meta.db" "$DATA_DIR/mailraven.db"

# Restore Blobs
mkdir -p "$DATA_DIR/blobs"
cp -r "$BACKUP_DIR/blobs/"* "$DATA_DIR/blobs/" || echo "No blobs to restore or copy failed (ignore if empty)"

echo "Restore complete. Please restart the server."
