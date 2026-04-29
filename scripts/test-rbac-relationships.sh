#!/bin/bash

# RBAC Test Script for Relationship Instances
# Tests permission-based access control for relationship endpoints
# Usage: ./scripts/test-rbac-relationships.sh [--keep-data]

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

# Test data
SOURCE_PUBLIC_ID=""
TARGET_PUBLIC_ID=""
RELATIONSHIP_PUBLIC_ID=""

# Parse arguments
for arg in "$@"; do
    case $arg in
        --keep-data)
            KEEP_DATA=true
            ;;
    esac
done

echo -e "${YELLOW}=== RBAC Relationship Instances Tests ===${NC}"
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

# Get seeded test object public_ids from database
get_test_object_ids() {
    echo -e "\n${BLUE}Fetching seeded test object IDs...${NC}"

    SOURCE_PUBLIC_ID=$(docker exec service-boilerplate-postgres psql -U postgres -d service_db -t -c "
        SELECT public_id FROM objects_service.objects WHERE name = 'Test Portfolio A' LIMIT 1;" 2>/dev/null | xargs)

    TARGET_PUBLIC_ID=$(docker exec service-boilerplate-postgres psql -U postgres -d service_db -t -c "
        SELECT public_id FROM objects_service.objects WHERE name = 'Test Asset X' LIMIT 1;" 2>/dev/null | xargs)

    if [ -z "$SOURCE_PUBLIC_ID" ] || [ -z "$TARGET_PUBLIC_ID" ]; then
        echo -e "${RED}✗ Could not find seeded test objects${NC}"
        echo "SOURCE_PUBLIC_ID: $SOURCE_PUBLIC_ID"
        echo "TARGET_PUBLIC_ID: $TARGET_PUBLIC_ID"
        exit 1
    fi

    echo -e "${GREEN}✓ Source: $SOURCE_PUBLIC_ID${NC}"
    echo -e "${GREEN}✓ Target: $TARGET_PUBLIC_ID${NC}"
}

# Test Relationships - CRUD
test_relationships() {
    local test_type_key="contains"

    echo -e "\n${YELLOW}=== Testing Relationships Permissions ===${NC}"

    # RL-1: object.admin CREATE relationship instance → 201
    echo -e "\n${BLUE}--- RL-1: Create Relationship (object.admin) ---${NC}"
    create_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/relationships" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $OBJECT_ADMIN_TOKEN" \
        -d "{\"source_object_id\": \"$SOURCE_PUBLIC_ID\", \"target_object_id\": \"$TARGET_PUBLIC_ID\", \"type_key\": \"$test_type_key\", \"status\": \"active\"}")

    if echo "$create_response" | jq -e '.public_id' >/dev/null 2>&1; then
        RELATIONSHIP_PUBLIC_ID=$(echo "$create_response" | jq -r '.public_id')
        echo -e "${GREEN}✓ Relationship created: $RELATIONSHIP_PUBLIC_ID${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ Failed to create relationship${NC}"
        echo "$create_response" | jq . 2>/dev/null || echo "$create_response"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi

    # RL-2: test.user CREATE relationship → 403
    make_request "POST" "$API_GATEWAY_URL/api/v1/relationships" "$TEST_USER_TOKEN" \
        "{\"source_object_id\": \"$SOURCE_PUBLIC_ID\", \"target_object_id\": \"$TARGET_PUBLIC_ID\", \"relationship_type_key\": \"$test_type_key\", \"status\": \"active\"}" "403" \
        "RL-2: Create Relationship (test.user) - should be 403"

    # RL-3: test.user LIST relationships → 200
    make_request "GET" "$API_GATEWAY_URL/api/v1/relationships" "$TEST_USER_TOKEN" \
        "" "200" "RL-3: List Relationships (test.user)"

    # RL-4: test.user GET specific relationship → 200
    make_request "GET" "$API_GATEWAY_URL/api/v1/relationships/$RELATIONSHIP_PUBLIC_ID" "$TEST_USER_TOKEN" \
        "" "200" "RL-4: Get Relationship (test.user)"

    # RL-5: object.admin UPDATE relationship → 200
    make_request "PUT" "$API_GATEWAY_URL/api/v1/relationships/$RELATIONSHIP_PUBLIC_ID" "$OBJECT_ADMIN_TOKEN" \
        "{\"status\": \"inactive\", \"relationship_metadata\": {\"test\": \"updated\"}}" "200" \
        "RL-5: Update Relationship (object.admin)"

    # RL-6: test.user UPDATE relationship → 403
    make_request "PUT" "$API_GATEWAY_URL/api/v1/relationships/$RELATIONSHIP_PUBLIC_ID" "$TEST_USER_TOKEN" \
        "{\"status\": \"active\"}" "403" \
        "RL-6: Update Relationship (test.user) - should be 403"

    # RL-7: object.admin DELETE relationship → 204
    make_request "DELETE" "$API_GATEWAY_URL/api/v1/relationships/$RELATIONSHIP_PUBLIC_ID" "$OBJECT_ADMIN_TOKEN" \
        "" "204" "RL-7: Delete Relationship (object.admin)"

    # Create another relationship for object-level tests
    create_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/relationships" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $OBJECT_ADMIN_TOKEN" \
        -d "{\"source_object_id\": \"$SOURCE_PUBLIC_ID\", \"target_object_id\": \"$TARGET_PUBLIC_ID\", \"type_key\": \"$test_type_key\", \"status\": \"active\"}")

    if echo "$create_response" | jq -e '.public_id' >/dev/null 2>&1; then
        RELATIONSHIP_PUBLIC_ID=$(echo "$create_response" | jq -r '.public_id')
    else
        echo -e "${RED}✗ Failed to create relationship for object-level tests${NC}"
        return 1
    fi

    # RL-8: test.user DELETE relationship → 403
    make_request "DELETE" "$API_GATEWAY_URL/api/v1/relationships/$RELATIONSHIP_PUBLIC_ID" "$TEST_USER_TOKEN" \
        "" "403" "RL-8: Delete Relationship (test.user) - should be 403"

    # RL-9: test.user GET relationships for object → 200
    make_request "GET" "$API_GATEWAY_URL/api/v1/objects/public-id/$SOURCE_PUBLIC_ID/relationships" "$TEST_USER_TOKEN" \
        "" "200" "RL-9: Get Relationships for Object (test.user)"

    # RL-10: test.user GET relationships for object by type → 200
    make_request "GET" "$API_GATEWAY_URL/api/v1/objects/public-id/$SOURCE_PUBLIC_ID/relationships/$test_type_key" "$TEST_USER_TOKEN" \
        "" "200" "RL-10: Get Relationships for Object by Type (test.user)"
}

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}=== Cleaning Up Test Data ===${NC}"

    if [ "$KEEP_DATA" = true ]; then
        echo -e "${YELLOW}⚠ Keeping test data (--keep-data flag set)${NC}"
        return 0
    fi

    # Delete relationship instances created during tests
    relationships_deleted=$(docker exec service-boilerplate-postgres psql -U postgres -d service_db -t -c "
        DELETE FROM objects_service.objects_relationships
        WHERE source_object_id = (
            SELECT id FROM objects_service.objects WHERE public_id = '$SOURCE_PUBLIC_ID'
        )
        AND target_object_id = (
            SELECT id FROM objects_service.objects WHERE public_id = '$TARGET_PUBLIC_ID'
        )
        AND relationship_type_id = (
            SELECT id FROM objects_service.objects_relationship_types WHERE type_key = 'contains'
        )
        RETURNING id;" 2>&1 | wc -l)

    echo -e "${GREEN}✓ Deleted $relationships_deleted test relationships${NC}"
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

    # Get test object IDs from seeded data
    get_test_object_ids

    # Run tests
    test_relationships

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