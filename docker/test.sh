#!/bin/bash
# API Test Script

API_URL="${API_URL:-http://localhost:8000}"
API_KEY="${API_KEY:-test123456789}"

echo "========================================"
echo "rsyslog REST API - Test Suite"
echo "========================================"
echo "API URL: $API_URL"
echo "API Key: ${API_KEY:0:20}..."
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0

# Test function
test_endpoint() {
    local name="$1"
    local endpoint="$2"
    local use_auth="${3:-yes}"
    
    echo -n "[$((PASSED + FAILED + 1))] $name... "
    
    if [ "$use_auth" = "yes" ]; then
        http_code=$(curl -s -w "%{http_code}" -o /tmp/response.json \
          -H "X-API-Key: $API_KEY" "$API_URL$endpoint")
    else
        http_code=$(curl -s -w "%{http_code}" -o /tmp/response.json "$API_URL$endpoint")
    fi
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ OK${NC} (HTTP $http_code)"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAILED${NC} (HTTP $http_code)"
        cat /tmp/response.json
        echo ""
        FAILED=$((FAILED + 1))
    fi
}

# Wait for API to be ready
echo -n "Waiting for API... "
for i in {1..30}; do
    if curl -s "$API_URL/health" > /dev/null 2>&1; then
        echo -e "${GREEN}ready${NC}"
        break
    fi
    [ $i -eq 30 ] && echo -e "${RED}timeout${NC}" && exit 1
    sleep 1
done
echo ""

# Run tests
echo "Running tests..."
echo ""

test_endpoint "Health Check" "/health" "no"
test_endpoint "Get Logs (default)" "/logs" "yes"
test_endpoint "Get Logs (limit 5)" "/logs?limit=5" "yes"
test_endpoint "Get Logs (limit 1)" "/logs?limit=1" "yes"
test_endpoint "Filter by Priority 3" "/logs?Priority=3" "yes"
test_endpoint "Filter by Priority 6" "/logs?Priority=6" "yes"
test_endpoint "Filter by Facility 1" "/logs?Facility=1" "yes"
test_endpoint "Filter by FromHost" "/logs?FromHost=webserver01" "yes"
test_endpoint "Filter by Message" "/logs?Message=login" "yes"
test_endpoint "Get Meta (list)" "/meta" "yes"
test_endpoint "Get Meta FromHost" "/meta/FromHost" "yes"
test_endpoint "Get Meta Priority" "/meta/Priority" "yes"
test_endpoint "Get Meta SysLogTag" "/meta/SysLogTag" "yes"
test_endpoint "Meta with filter" "/meta/SysLogTag?FromHost=webserver01" "yes"

echo ""
echo "========================================"
echo "Test Summary"
echo "========================================"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo "Total:  $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed! ✗${NC}"
    exit 1
fi
