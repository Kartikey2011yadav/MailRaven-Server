#!/bin/bash
set -e

echo "Running integration tests in Docker container..."

# We use 'docker compose run' instead of 'up' to avoid starting the main application server.
# This saves memory (no backend service daemon) and ensures clean test isolation.
# We also use --rm to clean up the container after.
# We explicitly do NOT start 'db' because the tests use internal SQLite.
# If tests ever require Postgres, add the dependency back or use 'run --service-ports'.

docker compose -f docker-compose.dev.yml run --rm --no-deps backend go test ./tests/... -v -p 1

echo "Tests completed."
