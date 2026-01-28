#!/bin/bash
set -e

echo "Starting Docker environment for testing..."
docker-compose -f docker-compose.dev.yml up -d db backend

echo "Waiting for database to be ready..."
# Simple wait loop for postgres
for i in {1..30}; do
  if docker exec mailraven-db-dev pg_isready -U mailraven > /dev/null 2>&1; then
    echo "Database is ready!"
    break
  fi
  echo "Waiting for DB..."
  sleep 2
done

echo "Running integration tests inside container..."
docker exec mailraven-backend-dev go test ./tests/... -v

echo "Cleaning up..."
docker-compose -f docker-compose.dev.yml down
