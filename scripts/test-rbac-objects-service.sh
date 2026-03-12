#!/bin/bash

# RBAC Test Script for Objects-Service
# Tests permission-based access control for objects-service endpoints
# Usage: ./scripts/test-rbac-objects-service.sh [--keep-data]

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

echo -e "${YELLOW}=== RBAC Objects-Service Tests ===${NC}"
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

# Test Object Types - CRUD
test_object_types() {
    local type_name="RBAC Test Type $TIMESTAMP"
    local type_id=""

    echo -e "\n${YELLOW}=== Testing Object Types Permissions ===${NC}"

    # OT-1: object.admin CREATE object type → 201
    echo -e "\n${BLUE}--- OT-1: Create Object Type (object.admin) ---${NC}"
    create_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/object-types" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $OBJECT_ADMIN_TOKEN" \
        -d "{\"name\": \"$type_name\", \"description\": \"Test object type for RBAC\"}")

    if echo "$create_response" | jq -e '.data.id' >/dev/null 2>&1; then
        type_id=$(echo "$create_response" | jq -r '.data.id')
        echo -e "${GREEN}✓ Object Type created with ID: $type_id${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ Failed to create object type${NC}"
        echo "$create_response" | jq . 2>/dev/null || echo "$create_response"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi

    # OT-2: test.user CREATE object type → 403
    make_request "POST" "$API_GATEWAY_URL/api/v1/object-types" "$TEST_USER_TOKEN" \
        "{\"name\": \"Unauthorized Type\", \"description\": \"Should fail\"}" "403" \
        "OT-2: Create Object Type (test.user) - should be 403"

    # OT-3: test.user READ object types → 200
    make_request "GET" "$API_GATEWAY_URL/api/v1/object-types" "$TEST_USER_TOKEN" \
        "" "200" "OT-3: Read Object Types (test.user)"

    # OT-4: object.admin UPDATE object type → 200
    make_request "PUT" "$API_GATEWAY_URL/api/v1/object-types/$type_id" "$OBJECT_ADMIN_TOKEN" \
        "{\"description\": \"Updated description\"}" "200" \
        "OT-4: Update Object Type (object.admin)"

    # OT-5: test.user UPDATE object type → 403
    make_request "PUT" "$API_GATEWAY_URL/api/v1/object-types/$type_id" "$TEST_USER_TOKEN" \
        "{\"description\": \"Should fail\"}" "403" \
        "OT-5: Update Object Type (test.user) - should be 403"

    # OT-6: object.admin DELETE object type → 204
    make_request "DELETE" "$API_GATEWAY_URL/api/v1/object-types/$type_id" "$OBJECT_ADMIN_TOKEN" \
        "" "204" "OT-6: Delete Object Type (object.admin)"

    # OT-7: test.user DELETE object type → 403 (type already deleted, skip)
    # For completeness, we test with a seeded type but expect 403
    make_request "DELETE" "$API_GATEWAY_URL/api/v1/object-types/1" "$TEST_USER_TOKEN" \
        "" "403" "OT-7: Delete Object Type (test.user) - should be 403"

    echo "$type_id"
}

# Test Objects - CRUD + Ownership
test_objects() {
    local object_type_id=""
    local test_user_object_id=""
    local admin_object_id=""

    echo -e "\n${YELLOW}=== Testing Objects Permissions ===${NC}"

    # Create object type for objects tests
    type_name="RBAC Objects Test Type $TIMESTAMP"
    create_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/object-types" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $OBJECT_ADMIN_TOKEN" \
        -d "{\"name\": \"$type_name\", \"description\": \"Test type for RBAC objects\"}")

    if echo "$create_response" | jq -e '.data.id' >/dev/null 2>&1; then
        object_type_id=$(echo "$create_response" | jq -r '.data.id')
        echo -e "${GREEN}✓ Created test object type: $object_type_id${NC}"
    else
        echo -e "${RED}✗ Failed to create test object type${NC}"
        echo "$create_response" | jq . 2>/dev/null || echo "$create_response"
        exit 1
    fi

    # Create object as test.user
    object_a_name="Test User Object $TIMESTAMP"
    obj_a_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/objects" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TEST_USER_TOKEN" \
        -d "{\"object_type_id\": $object_type_id, \"name\": \"$object_a_name\", \"description\": \"Object owned by test.user\"}")

    if echo "$obj_a_response" | jq -e '.data.id' >/dev/null 2>&1; then
        test_user_object_id=$(echo "$obj_a_response" | jq -r '.data.id')
        echo -e "${GREEN}✓ Created test object (test.user): $test_user_object_id${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))  # OBJ-1 passed
    else
        echo -e "${RED}✗ Failed to create test object${NC}"
        echo "$obj_a_response" | jq . 2>/dev/null || echo "$obj_a_response"
        exit 1
    fi

    # Create object as admin
    object_b_name="Admin Object $TIMESTAMP"
    obj_b_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/objects" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d "{\"object_type_id\": $object_type_id, \"name\": \"$object_b_name\", \"description\": \"Object owned by admin\"}")

    if echo "$obj_b_response" | jq -e '.data.id' >/dev/null 2>&1; then
        admin_object_id=$(echo "$obj_b_response" | jq -r '.data.id')
        echo -e "${GREEN}✓ Created test object (admin): $admin_object_id${NC}"
    else
        echo -e "${RED}✗ Failed to create admin object${NC}"
        echo "$obj_b_response" | jq . 2>/dev/null || echo "$obj_b_response"
        exit 1
    fi

    # OBJ-2: test.user READ own object → 200
    make_request "GET" "$API_GATEWAY_URL/api/v1/objects/$test_user_object_id" "$TEST_USER_TOKEN" \
        "" "200" "OBJ-2: Read Own Object (test.user)"

    # OBJ-3: test.user READ admin's object → 403
    make_request "GET" "$API_GATEWAY_URL/api/v1/objects/$admin_object_id" "$TEST_USER_TOKEN" \
        "" "403" "OBJ-3: Read Other's Object (test.user)"

    # OBJ-4: object.admin READ all objects → 200
    make_request "GET" "$API_GATEWAY_URL/api/v1/objects/$admin_object_id" "$OBJECT_ADMIN_TOKEN" \
        "" "200" "OBJ-4: Read Any Object (object.admin)"

    # OBJ-5: test.user UPDATE own object → 200
    make_request "PUT" "$API_GATEWAY_URL/api/v1/objects/$test_user_object_id" "$TEST_USER_TOKEN" \
        "{\"description\": \"Updated by test.user\"}" "200" \
        "OBJ-5: Update Own Object (test.user)"

    # OBJ-6: test.user UPDATE admin's object → 403
    make_request "PUT" "$API_GATEWAY_URL/api/v1/objects/$admin_object_id" "$TEST_USER_TOKEN" \
        "{\"description\": \"Should fail\"}" "403" \
        "OBJ-6: Update Other's Object (test.user)"

    # OBJ-7: admin UPDATE any object → 200
    make_request "PUT" "$API_GATEWAY_URL/api/v1/objects/$test_user_object_id" "$ADMIN_TOKEN" \
        "{\"description\": \"Updated by admin\"}" "200" \
        "OBJ-7: Update Any Object (admin)"

    # OBJ-8: test.user DELETE own object → 204
    make_request "DELETE" "$API_GATEWAY_URL/api/v1/objects/$test_user_object_id" "$TEST_USER_TOKEN" \
        "" "204" "OBJ-8: Delete Own Object (test.user)"

    # OBJ-9: test.user DELETE admin's object → 403
    make_request "DELETE" "$API_GATEWAY_URL/api/v1/objects/$admin_object_id" "$TEST_USER_TOKEN" \
        "" "403" "OBJ-9: Delete Other's Object (test.user)"

    # OBJ-10: admin DELETE any object → 204
    make_request "DELETE" "$API_GATEWAY_URL/api/v1/objects/$admin_object_id" "$ADMIN_TOKEN" \
        "" "204" "OBJ-10: Delete Any Object (admin)"

    # Store object_type_id for metadata/tags tests
    echo "$object_type_id"
}

# Test Metadata Operations
test_metadata() {
    local object_type_id=$1

    echo -e "\n${YELLOW}=== Testing Metadata Operations ===${NC}"

    # Create objects for metadata tests
    test_obj_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/objects" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TEST_USER_TOKEN" \
        -d "{\"object_type_id\": $object_type_id, \"name\": \"Meta Test User $TIMESTAMP\"}")
    test_meta_obj_id=$(echo "$test_obj_response" | jq -r '.data.id')

    admin_obj_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/objects" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d "{\"object_type_id\": $object_type_id, \"name\": \"Meta Test Admin $TIMESTAMP\"}")
    admin_meta_obj_id=$(echo "$admin_obj_response" | jq -r '.data.id')

    echo -e "${GREEN}Created objects for metadata tests: user=$test_meta_obj_id, admin=$admin_meta_obj_id${NC}"

    # META-1: test.user UPDATE own metadata → 200
    make_request "PUT" "$API_GATEWAY_URL/api/v1/objects/$test_meta_obj_id/metadata" "$TEST_USER_TOKEN" \
        "{\"key\": \"test\"}" "200" "META-1: Update Own Metadata (test.user)"

    # META-2: test.user UPDATE admin's metadata → 403
    make_request "PUT" "$API_GATEWAY_URL/api/v1/objects/$admin_meta_obj_id/metadata" "$TEST_USER_TOKEN" \
        "{\"key\": \"test\"}" "403" "META-2: Update Other's Metadata (test.user)"

    # META-3: admin UPDATE any metadata → 200
    make_request "PUT" "$API_GATEWAY_URL/api/v1/objects/$test_meta_obj_id/metadata" "$ADMIN_TOKEN" \
        "{\"key\": \"admin\"}" "200" "META-3: Update Any Metadata (admin)"
}

# Test Tags Operations
test_tags() {
    local object_type_id=$1

    echo -e "\n${YELLOW}=== Testing Tags Operations ===${NC}"

    # Create objects for tags tests
    test_obj_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/objects" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TEST_USER_TOKEN" \
        -d "{\"object_type_id\": $object_type_id, \"name\": \"Tags Test User $TIMESTAMP\"}")
    test_tags_obj_id=$(echo "$test_obj_response" | jq -r '.data.id')

    admin_obj_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/objects" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d "{\"object_type_id\": $object_type_id, \"name\": \"Tags Test Admin $TIMESTAMP\"}")
    admin_tags_obj_id=$(echo "$admin_obj_response" | jq -r '.data.id')

    echo -e "${GREEN}Created objects for tags tests: user=$test_tags_obj_id, admin=$admin_tags_obj_id${NC}"

    # TAGS-1: test.user ADD own tags → 200
    make_request "POST" "$API_GATEWAY_URL/api/v1/objects/$test_tags_obj_id/tags" "$TEST_USER_TOKEN" \
        "{\"tags\": [\"test-tag\"]}" "200" "TAGS-1: Add Own Tags (test.user)"

    # TAGS-2: test.user ADD admin's tags → 403
    make_request "POST" "$API_GATEWAY_URL/api/v1/objects/$admin_tags_obj_id/tags" "$TEST_USER_TOKEN" \
        "{\"tags\": [\"test-tag\"]}" "403" "TAGS-2: Add Other's Tags (test.user)"

    # TAGS-3: admin ADD any tags → 200
    make_request "POST" "$API_GATEWAY_URL/api/v1/objects/$test_tags_obj_id/tags" "$ADMIN_TOKEN" \
        "{\"tags\": [\"admin-tag\"]}" "200" "TAGS-3: Add Any Tags (admin)"

    # TAGS-4: test.user DELETE own tags → 200
    make_request "DELETE" "$API_GATEWAY_URL/api/v1/objects/$test_tags_obj_id/tags" "$TEST_USER_TOKEN" \
        "{\"tags\": [\"test-tag\"]}" "200" "TAGS-4: Delete Own Tags (test.user)"

    # TAGS-5: test.user DELETE admin's tags → 403
    make_request "DELETE" "$API_GATEWAY_URL/api/v1/objects/$admin_tags_obj_id/tags" "$TEST_USER_TOKEN" \
        "{\"tags\": [\"admin-tag\"]}" "403" "TAGS-5: Delete Other's Tags (test.user)"

    # TAGS-6: admin DELETE any tags → 200
    make_request "DELETE" "$API_GATEWAY_URL/api/v1/objects/$test_tags_obj_id/tags" "$ADMIN_TOKEN" \
        "{\"tags\": [\"admin-tag\"]}" "200" "TAGS-6: Delete Any Tags (admin)"
}

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}=== Cleaning Up Test Data ===${NC}"

    if [ "$KEEP_DATA" = true ]; then
        echo -e "${YELLOW}⚠ Keeping test data (--keep-data flag set)${NC}"
        return 0
    fi

    # Delete objects with test timestamp in name
    objects_deleted=$(docker exec service-boilerplate-postgres psql -U postgres -d service_db -t -c "
        DELETE FROM objects_service.objects 
        WHERE name LIKE '%$TIMESTAMP%'
        RETURNING id;" 2>&1 | wc -l)

    echo -e "${GREEN}✓ Deleted $objects_deleted test objects${NC}"

    # Delete object types with test timestamp in name
    types_deleted=$(docker exec service-boilerplate-postgres psql -U postgres -d service_db -t -c "
        DELETE FROM objects_service.object_types 
        WHERE name LIKE '%$TIMESTAMP%'
        RETURNING id;" 2>&1 | wc -l)

    echo -e "${GREEN}✓ Deleted $types_deleted test object types${NC}"
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
    test_object_types
    test_objects
    test_metadata 1  # Use seeded type ID for metadata tests
    test_tags 1  # Use seeded type ID for tags tests

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
