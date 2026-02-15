#!/bin/bash
# verify-imports.sh
# Verifies that all import paths have been updated correctly

set -e

echo "=========================================="
echo "Verifying Import Paths"
echo "=========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check for old import paths
OLD_IMPORTS=$(grep -r "github.com/phil-bot/rsyslog-rest-api" --include="*.go" . 2>/dev/null || true)

if [ -n "$OLD_IMPORTS" ]; then
    echo "${RED}✗ Found old import paths:${NC}"
    echo "$OLD_IMPORTS"
    echo ""
    echo "${YELLOW}Please run: ./update-imports.sh${NC}"
    exit 1
fi

# Check for new import paths
NEW_IMPORTS=$(grep -r "github.com/phil-bot/rsyslox" --include="*.go" . 2>/dev/null || true)

if [ -z "$NEW_IMPORTS" ]; then
    echo "${YELLOW}⚠ No new import paths found${NC}"
    echo "This might be okay if your project doesn't use internal packages"
else
    IMPORT_COUNT=$(echo "$NEW_IMPORTS" | wc -l)
    echo "${GREEN}✓ All imports updated correctly${NC}"
    echo "Found $IMPORT_COUNT references to new import path"
fi

echo ""
echo "=========================================="
echo "${GREEN}✓ Verification Complete${NC}"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  1. Run: go mod tidy"
echo "  2. Run: go build"
echo "  3. Run: go test ./..."
