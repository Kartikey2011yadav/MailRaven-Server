Write-Host "MailRaven Setup" -ForegroundColor Green
Write-Host "=============================="

# 1. Check Dependencies
Write-Host "`n[1/4] Checking dependencies..." -ForegroundColor Green

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Error: Go is not installed." -ForegroundColor Red
    exit 1
}
Write-Host "Go version: $(go version)"

if (-not (Get-Command npm -ErrorAction SilentlyContinue)) {
    Write-Host "Error: npm is not installed (required for client)." -ForegroundColor Red
    exit 1
}
Write-Host "npm version: $(npm --version)"

# 2. Build Backend
Write-Host "`n[2/4] Building backend..." -ForegroundColor Green
$ErrorActionPreference = "Stop"

if (Test-Path "bin") { Remove-Item -Path "bin" -Recurse -Force -ErrorAction SilentlyContinue }
New-Item -ItemType Directory -Force -Path "bin" | Out-Null

Write-Host "Building Server..."
try {
    go build -o bin/mailraven.exe ./cmd/mailraven
    Write-Host "Building CLI..."
    go build -o bin/mailraven-cli.exe ./cmd/mailraven-cli
} catch {
    Write-Host "Build failed: $_" -ForegroundColor Red
    exit 1
}

# 3. Build Frontend
Write-Host "`n[3/4] Building frontend..." -ForegroundColor Green
if (Test-Path "client") {
    Push-Location "client"
    Write-Host "Installing frontend dependencies..."
    npm install --silent
    Write-Host "Building frontend assets..."
    npm run build
    Pop-Location
} else {
    Write-Host "Warning: client directory not found." -ForegroundColor Yellow
}

# 4. Configuration Setup
Write-Host "`n[4/4] Setting up configuration..." -ForegroundColor Green

if (-not (Test-Path "config.yaml")) {
    if (Test-Path "config.example.yaml") {
        Write-Host "Creating config.yaml from example..."
        Copy-Item "config.example.yaml" "config.yaml"
    } else {
        Write-Host "No config.example.yaml found." -ForegroundColor Yellow
    }
} else {
    Write-Host "config.yaml already exists."
}

Write-Host "`nSetup Complete!" -ForegroundColor Green
Write-Host "Run backend: .\bin\mailraven.exe"
