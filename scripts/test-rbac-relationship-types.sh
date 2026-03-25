#!/bin/bash

# RBAC Test Script for Relationship Types
# Tests permission-based access control for relationship-types endpoints
# Usage: ./scripts/test-rbac-relationship-types.sh [--keep-data]

set -e

# Configuration
API_GATEWAY_URL="${API_GATEWAY_URL:-http://localhost:8080}"

# Test users
ADMIN_EMAIL="dev.admin@example.com"
ADMIN_PASSWORD="devadmin123"
OBJECT_ADMIN_EMAIL="object.admin@example.com"
OBJECT_ADMIN_PASSWORD="devadmin123"
TEST_USER_EMAIL="test.user@example.com"
TEST_USER_PASSWORD="devadmin123"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Global variables
ADMIN_TOKEN=""
OBJECT_ADMIN_TOKEN=""
TEST_USER_TOKEN=""
TIMESTAMP=""
KEEP_DATA=false

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Parse arguments
for arg in "$@"; do
    case $arg in
        --keep-data)
            KEEP_DATA=true
            ;;
    esac
done

echo -e "${YELLOW}=== RBAC Relationship Types Tests ===${NC}"
echo "API Gateway: $API_GATEWAY_URL"
echo "Timestamp: $(date +"%Y-%m-%d-%H-%M-%S")"
TIMESTAMP=$(date +"%Y-%m-%d-%H-%M-%S")

# Function to login and get token
login() {
    local email=$1
    local password=$2
    local description=$3

    echo -e "\n${BLUE}Logging in as $description...${NC}" >&2

    login_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"$email\", \"password\": \"$password\"}")

    if echo "$login_response" | jq -e '.access_token' >/dev/null 2>&1; then
        token=$(echo "$login_response" | jq -r '.access_token')
        echo -e "${GREEN}✓ Login successful: $description${NC}" >&2
        printf "%s" "$token"
        return 0
    else
        echo -e "${RED}✗ Login failed for $description${NC}" >&2
        echo "Response: $login_response" >&2
        exit 1
    fi
}

# Make authenticated request
# Args: method, url, token, data, expected_status, description
make_request() {
    local method=$1
    local url=$2
    local token=$3
    local data=$4
    local expected_status=$5
    local description=$6

    echo -e "\n${YELLOW}Testing: $description${NC}"
    echo -e "${BLUE}$method $url${NC}"

    if [ -z "$data" ] || [ "$data" = "null" ]; then
        response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $token" 2>/dev/null)
    else
        response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $token" \
            -d "$data" 2>/dev/null)
    fi

    http_status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_STATUS:/d')

    if [ "$http_status" = "$expected_status" ]; then
        echo -e "${GREEN}✓ PASS ($expected_status)${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAIL (expected $expected_status, got $http_status)${NC}"
        if [ -n "$body" ] && [ "$body" != "null" ]; then
            echo "Response: $body" | jq . 2>/dev/null || echo "$body"
        fi
        ((TESTS_FAILED++))
        return 1
    fi
}

# Test Relationship Types - CRUD
test_relationship_types() {
    local test_type_key="rbac_test_type_$TIMESTAMP"

    echo -e "\n${YELLOW}=== Testing Relationship Types Permissions ===${NC}"

    # RT-1: object.admin CREATE relationship type → 201
    echo -e "\n${BLUE}--- RT-1: Create Relationship Type (object.admin) ---${NC}"
    create_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/relationship-types" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $OBJECT_ADMIN_TOKEN" \
        -d "{\"type_key\": \"$test_type_key\", \"relationship_name\": \"RBAC Test Type\", \"cardinality\": \"one_to_one\"}")

    if echo "$create_response" | jq -e '.type_key' >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Relationship Type created: $test_type_key${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ Failed to create relationship type${NC}"
        echo "$create_response" | jq . 2>/dev/null || echo "$create_response"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi

    # RT-2: test.user CREATE relationship type → 403
    make_request "POST" "$API_GATEWAY_URL/api/v1/relationship-types" "$TEST_USER_TOKEN" \
        "{\"type_key\": \"unauthorized_type_$TIMESTAMP\", \"cardinality\": \"one_to_one\"}" "403" \
        "RT-2: Create Relationship Type (test.user) - should be 403"

    # RT-3: test.user READ relationship types → 200
    make_request "GET" "$API_GATEWAY_URL/api/v1/relationship-types" "$TEST_USER_TOKEN" \
        "" "200" "RT-3: Read Relationship Types (test.user)"

    # RT-4: test.user GET specific relationship type → 200
    make_request "GET" "$API_GATEWAY_URL/api/v1/relationship-types/contains" "$TEST_USER_TOKEN" \
        "" "200" "RT-4: Get Relationship Type (test.user)"

    # RT-5: object.admin UPDATE relationship type → 200
    make_request "PUT" "$API_GATEWAY_URL/api/v1/relationship-types/$test_type_key" "$OBJECT_ADMIN_TOKEN" \
        "{\"relationship_name\": \"Updated RBAC Test Type\", \"required\": true}" "200" \
        "RT-5: Update Relationship Type (object.admin)"

    # RT-6: test.user UPDATE relationship type → 403
    make_request "PUT" "$API_GATEWAY_URL/api/v1/relationship-types/$test_type_key" "$TEST_USER_TOKEN" \
        "{\"relationship_name\": \"Should fail\"}" "403" \
        "RT-6: Update Relationship Type (test.user) - should be 403"

    # RT-7: object.admin DELETE relationship type → 204
    make_request "DELETE" "$API_GATEWAY_URL/api/v1/relationship-types/$test_type_key" "$OBJECT_ADMIN_TOKEN" \
        "" "204" "RT-7: Delete Relationship Type (object.admin)"

    # RT-8: test.user DELETE relationship type → 403
    make_request "DELETE" "$API_GATEWAY_URL/api/v1/relationship-types/contains" "$TEST_USER_TOKEN" \
        "" "403" "RT-8: Delete Relationship Type (test.user) - should be 403"
}

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}=== Cleaning Up Test Data ===${NC}"

    if [ "$KEEP_DATA" = true ]; then
        echo -e "${YELLOW}⚠ Keeping test data (--keep-data flag set)${NC}"
        return 0
    fi

    # Delete relationship types with test timestamp in type_key
    types_deleted=$(docker exec service-boilerplate-postgres psql -U postgres -d service_db -t -c "
        DELETE FROM objects_service.objects_relationship_types ort
        USING objects_service.objects o
        WHERE o.id = ort.object_id
        AND o.name LIKE '%$TIMESTAMP%'
        RETURNING ort.type_key;" 2>&1 | wc -l)

    echo -e "${GREEN}✓ Deleted $types_deleted test relationship types${NC}"

    # Delete associated objects (should cascade, but doing explicitly for cleanup)
    objects_deleted=$(docker exec service-boilerplate-postgres psql -U postgres -d service_db -t -c "
        DELETE FROM objects_service.objects 
        WHERE name LIKE '%$TIMESTAMP%'
        RETURNING id;" 2>&1 | wc -l)

    echo -e "${GREEN}✓ Deleted $objects_deleted test objects${NC}"
}

# Main function
main() {
    # Check if jq is available
    if ! command -v jq &>/dev/null; then
        echo -e "${RED}✗ jq is required for this script. Please install jq.${NC}"
        exit 1
    fi

    # Login all users
    echo -e "\n${YELLOW}=== Logging In Test Users ===${NC}"
    ADMIN_TOKEN=$(login "$ADMIN_EMAIL" "$ADMIN_PASSWORD" "dev.admin")
    OBJECT_ADMIN_TOKEN=$(login "$OBJECT_ADMIN_EMAIL" "$OBJECT_ADMIN_PASSWORD" "object.admin")
    TEST_USER_TOKEN=$(login "$TEST_USER_EMAIL" "$TEST_USER_PASSWORD" "test.user")

    # Run tests
    test_relationship_types

    # Cleanup (unless --keep-data)
    cleanup

    # Print summary
    echo -e "\n${YELLOW}=== Test Summary ===${NC}"
    echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Failed: ${RED}$TESTS_FAILED${NC}"

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "\n${GREEN}🎉 All tests passed!${NC}"
        exit 0
    else
        echo -e "\n${RED}✗ Some tests failed${NC}"
        exit 1
    fi
}

# Run main
main "$@"