#!/bin/bash
set -e

# MailRaven Setup Script
# Usage: ./scripts/setup.sh [--prod]

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}MailRaven Setup${NC}"
echo "=============================="

# 1. Check Dependencies
echo -e "\n${GREEN}[1/4] Checking dependencies...${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed.${NC}"
    exit 1
fi
echo "Go version: $(go version)"

if ! command -v npm &> /dev/null; then
    echo -e "${RED}Error: npm is not installed (required for client).${NC}"
    exit 1
fi
echo "npm version: $(npm --version)"

# 2. Build Backend
echo -e "\n${GREEN}[2/4] Building backend...${NC}"
make build || {
    echo "Makefile failed or missing, trying manual build..."
    mkdir -p bin
    go build -o bin/mailraven ./cmd/mailraven
    go build -o bin/mailraven-cli ./cmd/mailraven-cli
}

# 3. Build Frontend
echo -e "\n${GREEN}[3/4] Building frontend...${NC}"
if [ -d "client" ]; then
    cd client
    echo "Installing frontend dependencies..."
    npm install --silent
    echo "Building frontend assets..."
    npm run build
    cd ..
else
    echo -e "${RED}Warning: client directory not found.${NC}"
fi

# 4. Configuration Setup
echo -e "\n${GREEN}[4/4] Setting up configuration...${NC}"
CONFIG_DIR="config" # Or /etc/mailraven
mkdir -p "$CONFIG_DIR"

if [ ! -f "config.yaml" ]; then
    if [ -f "config.example.yaml" ]; then
        echo "Creating config.yaml from example..."
        cp config.example.yaml config.yaml
    else
        echo "No config.example.yaml found. Please create config.yaml manually."
    fi
else
    echo "config.yaml already exists."
fi

# Generate Keys if needed
echo "Checking for DKIM/TLS keys..."
# We can run the genkeys script if needed, but usually handled by admin.
# go run scripts/genkeys.go --check-only

echo -e "\n${GREEN}Setup Complete!${NC}"
echo "Run backend: ./bin/mailraven"
echo "Run frontend: (served via backend if configured, or use 'cd client && npm run dev')"
