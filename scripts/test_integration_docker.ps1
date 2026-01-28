Write-Host "Starting Docker environment for testing..."
docker-compose -f docker-compose.dev.yml up -d db backend

Write-Host "Waiting for database to be ready..."
for ($i=0; $i -lt 30; $i++) {
    $result = docker exec mailraven-db-dev pg_isready -U mailraven 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Database is ready!"
        break
    }
    Write-Host "Waiting for DB..."
    Start-Sleep -Seconds 2
}

Write-Host "Running integration tests inside container..."
docker exec mailraven-backend-dev go test ./tests/... -v

Write-Host "Cleaning up..."
docker-compose -f docker-compose.dev.yml down
