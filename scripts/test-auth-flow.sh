#!/bin/bash

# Test script for end-to-end authentication flow
# Tests the complete authentication journey: register -> login -> access protected resource -> logout

set -e

# Configuration - can be overridden via environment variables
API_GATEWAY_URL="${API_GATEWAY_URL:-http://localhost:8080}"
USER_SERVICE_URL="${USER_SERVICE_URL:-http://localhost:8081}"
AUTH_SERVICE_URL="${AUTH_SERVICE_URL:-http://localhost:8083}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Testing End-to-End Authentication Flow${NC}"
echo "=========================================="

# Function to make HTTP requests and check response
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local auth_header=$4
    local expected_status=$5

    echo -e "\n${YELLOW}Testing: $method $url${NC}"

    if [ -n "$auth_header" ]; then
        response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -H "Authorization: $auth_header" \
            -d "$data" 2>/dev/null)
    else
        response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -d "$data" 2>/dev/null)
    fi

    http_status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_STATUS:/d')

    if [ "$http_status" = "$expected_status" ]; then
        echo -e "${GREEN}✓ Expected status $expected_status received${NC}"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 0
    else
        echo -e "${RED}✗ Expected status $expected_status, got $http_status${NC}"
        echo "Response: $body"
        return 1
    fi
}

# Test 1: Health checks
echo -e "\n${YELLOW}Step 1: Health Checks${NC}"
make_request "GET" "$API_GATEWAY_URL/health" "" "" "200"
make_request "GET" "$USER_SERVICE_URL/health" "" "" "200"
make_request "GET" "$AUTH_SERVICE_URL/health" "" "" "200"

# Test 2: User registration
echo -e "\n${YELLOW}Step 2: User Registration${NC}"
TIMESTAMP=$(date +%s)
register_data='{
    "email": "test-'$TIMESTAMP'@example.com",
    "password": "testpassword123",
    "first_name": "Test",
    "last_name": "User"
}'
make_request "POST" "$API_GATEWAY_URL/api/v1/auth/register" "$register_data" "" "201"

# Test 3: User login
echo -e "\n${YELLOW}Step 3: User Login${NC}"
login_data='{
    "email": "test-'$TIMESTAMP'@example.com",
    "password": "testpassword123"
}'
login_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "$login_data")

if echo "$login_response" | jq -e '.access_token' >/dev/null 2>&1; then
    echo -e "${GREEN}✓ Login successful${NC}"
    TOKEN=$(echo "$login_response" | jq -r '.access_token')
    echo "Token: ${TOKEN:0:50}..."
else
    echo -e "${RED}✗ Login failed${NC}"
    echo "Response: $login_response"
    exit 1
fi

# Test 4: Update user profile (PATCH)
echo -e "\n${YELLOW}Step 4: Update User Profile (PATCH)${NC}"
# Get user ID from login response
USER_ID=$(echo "$login_response" | jq -r '.user.id')
update_data='{
    "first_name": "Updated",
    "last_name": "User Name"
}'
make_request "PATCH" "$API_GATEWAY_URL/api/v1/users/$USER_ID" "$update_data" "Bearer $TOKEN" "200"

# Test 5: Access protected resource through API gateway
echo -e "\n${YELLOW}Step 5: Access Protected Resource via API Gateway${NC}"
make_request "GET" "$API_GATEWAY_URL/api/v1/users" "" "Bearer $TOKEN" "200"

# Test 6: Access user profile
echo -e "\n${YELLOW}Step 6: Access User Profile${NC}"
make_request "GET" "$API_GATEWAY_URL/api/v1/auth/me" "" "Bearer $TOKEN" "200"

# Test 7: Test unauthorized access
echo -e "\n${YELLOW}Step 7: Test Unauthorized Access${NC}"
make_request "GET" "$API_GATEWAY_URL/api/v1/users" "" "" "401"

# Test 8: Logout
echo -e "\n${YELLOW}Step 8: Logout${NC}"
make_request "POST" "$API_GATEWAY_URL/api/v1/auth/logout" "" "Bearer $TOKEN" "200"

# Test 9: Verify token is invalidated
echo -e "\n${YELLOW}Step 9: Verify Token Invalidated${NC}"
make_request "GET" "$API_GATEWAY_URL/api/v1/users" "" "Bearer $TOKEN" "401"

echo -e "\n${GREEN}🎉 All authentication tests passed!${NC}"
echo -e "${YELLOW}End-to-End Authentication Flow Test Complete${NC}"