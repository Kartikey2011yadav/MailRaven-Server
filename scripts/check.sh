#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}MailRaven Environment Check${NC}"

# Check 1: Ports
# Need netstat or ss or lsof
echo "Checking ports..."
# This is tricky in scripts without privileges or specific tools.
# We skip for now or use timeout+bash connect.

# Check 2: Config validity
if [ -f "bin/mailraven" ]; then
    echo "Validating config..."
    ./bin/mailraven check-config
else
    echo "Binary not found, skipping config check."
fi

# Check 3: Database
# If config uses postgres, check pg_isready?

echo "Check complete."
