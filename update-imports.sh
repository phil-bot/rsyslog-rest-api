#!/bin/bash
# update-imports.sh
# Updates all Go import paths from rsyslog-rest-api to rsyslox

set -e

echo "=========================================="
echo "Updating Go Import Paths"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Count files
TOTAL_FILES=0
UPDATED_FILES=0

# Find all .go files and update imports
echo "${BLUE}Searching for Go files...${NC}"
while IFS= read -r file; do
    TOTAL_FILES=$((TOTAL_FILES + 1))
    
    # Check if file contains old import path
    if grep -q "github.com/phil-bot/rsyslog-rest-api" "$file"; then
        echo "  Updating: $file"
        
        # Update the import path
        sed -i 's|github.com/phil-bot/rsyslog-rest-api|github.com/phil-bot/rsyslox|g' "$file"
        
        UPDATED_FILES=$((UPDATED_FILES + 1))
    fi
done < <(find . -name "*.go" -type f)

echo ""
echo "=========================================="
echo "${GREEN}✓ Update Complete${NC}"
echo "=========================================="
echo "Total Go files found: $TOTAL_FILES"
echo "Files updated: $UPDATED_FILES"
echo ""

if [ $UPDATED_FILES -gt 0 ]; then
    echo "Next steps:"
    echo "  1. Run: go mod tidy"
    echo "  2. Run: go build"
    echo "  3. Run: go test ./..."
fi
