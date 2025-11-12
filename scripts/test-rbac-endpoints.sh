#!/bin/bash

# Test script for RBAC (Role-Based Access Control) endpoints
# Tests role/permission management, user-role assignments, and access control

set -e

# Configuration
API_GATEWAY_URL="http://localhost:8080"
ADMIN_EMAIL="dev.admin@example.com" # Admin user we set up
ADMIN_PASSWORD="devadmin123"        # Password we configured

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${YELLOW}Testing RBAC (Role-Based Access Control) Endpoints${NC}"
echo "======================================================"

# Global variables
ADMIN_TOKEN=""
TEST_ROLE_ID=""
TEST_PERMISSION_ID=""

# Function to login as admin and get token
login_admin() {
    echo -e "\n${BLUE}Authenticating as admin user...${NC}"

    login_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"$ADMIN_EMAIL\", \"password\": \"$ADMIN_PASSWORD\"}")

    if echo "$login_response" | jq -e '.access_token' >/dev/null 2>&1; then
        ADMIN_TOKEN=$(echo "$login_response" | jq -r '.access_token')
        echo -e "${GREEN}✓ Admin login successful${NC}"
        return 0
    else
        echo -e "${RED}✗ Admin login failed${NC}"
        echo "Response: $login_response"
        exit 1
    fi
}

# Enhanced make_request function with RBAC-specific features
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    local description=$5

    echo -e "\n${YELLOW}Testing: $description${NC}"
    echo -e "${BLUE}$method $url${NC}"

    response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X "$method" "$url" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d "$data" 2>/dev/null)

    http_status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_STATUS:/d')

    if [ "$http_status" = "$expected_status" ]; then
        echo -e "${GREEN}✓ Expected status $expected_status received${NC}"
        if [ -n "$body" ] && [ "$body" != "null" ]; then
            echo "$body" | jq . 2>/dev/null || echo "$body"
        fi
        return 0
    else
        echo -e "${RED}✗ Expected status $expected_status, got $http_status${NC}"
        if [ -n "$body" ] && [ "$body" != "null" ]; then
            echo "Response: $body"
        fi
        return 1
    fi
}

# Test functions for different RBAC operations
test_role_management() {
    echo -e "\n${YELLOW}=== Testing Role Management ===${NC}"

    # Create a test role
    echo -e "\n${BLUE}Creating test role...${NC}"
    create_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/auth/roles" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d '{"name": "test_manager", "description": "Test manager role for RBAC testing"}')

    if echo "$create_response" | jq -e '.id' >/dev/null 2>&1; then
        TEST_ROLE_ID=$(echo "$create_response" | jq -r '.id')
        echo -e "${GREEN}✓ Test role created with ID: $TEST_ROLE_ID${NC}"
        echo "$create_response" | jq . 2>/dev/null || echo "$create_response"
    else
        echo -e "${RED}✗ Failed to create test role${NC}"
        echo "Response: $create_response"
        return 1
    fi

    # List all roles
    make_request "GET" "$API_GATEWAY_URL/api/v1/auth/roles" \
        "" "200" "List all roles"

    # Get specific role
    if [ -n "$TEST_ROLE_ID" ]; then
        make_request "GET" "$API_GATEWAY_URL/api/v1/auth/roles/$TEST_ROLE_ID" \
            "" "200" "Get specific test role"

        # Update role
        make_request "PUT" "$API_GATEWAY_URL/api/v1/auth/roles/$TEST_ROLE_ID" \
            '{"name": "test_manager", "description": "Updated test manager role"}' \
            "200" "Update test role"
    fi
}

test_permission_management() {
    echo -e "\n${YELLOW}=== Testing Permission Management ===${NC}"

    # List permissions
    list_perm_response=$(curl -s -X GET "$API_GATEWAY_URL/api/v1/auth/permissions" \
        -H "Authorization: Bearer $ADMIN_TOKEN")

    if echo "$list_perm_response" | jq -e '.permissions[0].id' >/dev/null 2>&1; then
        TEST_PERMISSION_ID=$(echo "$list_perm_response" | jq -r '.permissions[0].id')
        echo -e "${GREEN}✓ Found permission with ID: $TEST_PERMISSION_ID${NC}"
        echo "$list_perm_response" | jq '.permissions[0]' 2>/dev/null || echo "First permission found"
    else
        echo -e "${RED}✗ Failed to list permissions${NC}"
        echo "Response: $list_perm_response"
        return 1
    fi

    # Create a test permission
    echo -e "\n${BLUE}Creating test permission...${NC}"
    create_perm_response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/auth/permissions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d '{"name": "test:access", "resource": "test", "action": "access"}')

    if echo "$create_perm_response" | jq -e '.id' >/dev/null 2>&1; then
        TEST_PERMISSION_ID=$(echo "$create_perm_response" | jq -r '.id')
        echo -e "${GREEN}✓ Test permission created with ID: $TEST_PERMISSION_ID${NC}"
        echo "$create_perm_response" | jq . 2>/dev/null || echo "$create_perm_response"
    else
        echo -e "${RED}✗ Failed to create test permission${NC}"
        echo "Response: $create_perm_response"
        return 1
    fi

    # Get specific permission
    make_request "GET" "$API_GATEWAY_URL/api/v1/auth/permissions/$TEST_PERMISSION_ID" \
        "" "200" "Get specific test permission"

    # Update permission
    make_request "PUT" "$API_GATEWAY_URL/api/v1/auth/permissions/$TEST_PERMISSION_ID" \
        '{"name": "test:access", "resource": "test", "action": "read"}' \
        "200" "Update test permission"
}

test_role_permission_relationships() {
    echo -e "\n${YELLOW}=== Testing Role-Permission Relationships ===${NC}"

    if [ -z "$TEST_ROLE_ID" ] || [ -z "$TEST_PERMISSION_ID" ]; then
        echo -e "${RED}✗ Missing test role or permission ID${NC}"
        return 1
    fi

    # Assign permission to role
    make_request "POST" "$API_GATEWAY_URL/api/v1/auth/roles/$TEST_ROLE_ID/permissions" \
        "{\"permission_id\": \"$TEST_PERMISSION_ID\"}" \
        "200" "Assign permission to role"

    # Get role permissions
    make_request "GET" "$API_GATEWAY_URL/api/v1/auth/roles/$TEST_ROLE_ID/permissions" \
        "" "200" "Get role permissions"

    # Remove permission from role
    make_request "DELETE" "$API_GATEWAY_URL/api/v1/auth/roles/$TEST_ROLE_ID/permissions/$TEST_PERMISSION_ID" \
        "" "200" "Remove permission from role"
}

test_user_role_management() {
    echo -e "\n${YELLOW}=== Testing User-Role Management ===${NC}"

    # Get a test user ID (using the admin user itself for testing)
    admin_user_response=$(curl -s -X GET "$API_GATEWAY_URL/api/v1/auth/me" \
        -H "Authorization: Bearer $ADMIN_TOKEN")

    if echo "$admin_user_response" | jq -e '.user.id' >/dev/null 2>&1; then
        TEST_USER_ID=$(echo "$admin_user_response" | jq -r '.user.id')
        echo -e "${GREEN}✓ Using admin user ID: $TEST_USER_ID for testing${NC}"
    else
        echo -e "${RED}✗ Failed to get admin user info${NC}"
        echo "Response: $admin_user_response"
        return 1
    fi

    # Get user roles
    make_request "GET" "$API_GATEWAY_URL/api/v1/auth/users/$TEST_USER_ID/roles" \
        "" "200" "Get user roles"

    if [ -n "$TEST_ROLE_ID" ]; then
        # Assign role to user
        make_request "POST" "$API_GATEWAY_URL/api/v1/auth/users/$TEST_USER_ID/roles" \
            "{\"role_id\": \"$TEST_ROLE_ID\"}" \
            "200" "Assign role to user"

        # Remove role from user
        make_request "DELETE" "$API_GATEWAY_URL/api/v1/auth/users/$TEST_USER_ID/roles/$TEST_ROLE_ID" \
            "" "200" "Remove role from user"

        # Bulk update user roles (restore original roles)
        make_request "PUT" "$API_GATEWAY_URL/api/v1/auth/users/$TEST_USER_ID/roles" \
            "{\"role_ids\": [\"d149a841-62fc-4256-83f6-819d08fa75cc\", \"c8e8af3b-612e-4c2f-b8a2-1c62950e7558\"]}" \
            "200" "Bulk update user roles"
    fi
}

cleanup_test_data() {
    echo -e "\n${YELLOW}=== Cleaning Up Test Data ===${NC}"

    # Delete test role
    if [ -n "$TEST_ROLE_ID" ]; then
        echo -e "${BLUE}Deleting test role...${NC}"
        delete_response=$(curl -s -X DELETE "$API_GATEWAY_URL/api/v1/auth/roles/$TEST_ROLE_ID" \
            -H "Authorization: Bearer $ADMIN_TOKEN")

        if echo "$delete_response" | jq -e '.message' >/dev/null 2>&1; then
            echo -e "${GREEN}✓ Test role deleted successfully${NC}"
        else
            echo -e "${YELLOW}⚠️  Could not delete test role (might be referenced elsewhere)${NC}"
        fi
    fi

    # Delete test permission
    if [ -n "$TEST_PERMISSION_ID" ]; then
        echo -e "${BLUE}Deleting test permission...${NC}"
        delete_perm_response=$(curl -s -X DELETE "$API_GATEWAY_URL/api/v1/auth/permissions/$TEST_PERMISSION_ID" \
            -H "Authorization: Bearer $ADMIN_TOKEN")

        if echo "$delete_perm_response" | jq -e '.message' >/dev/null 2>&1; then
            echo -e "${GREEN}✓ Test permission deleted successfully${NC}"
        else
            echo -e "${YELLOW}⚠️  Could not delete test permission (might be referenced elsewhere)${NC}"
        fi
    fi
}

# Main test execution
main() {
    # Check if jq is available
    if ! command -v jq &>/dev/null; then
        echo -e "${RED}✗ jq is required for this script. Please install jq.${NC}"
        exit 1
    fi

    # Login as admin first
    login_admin

    # Run all RBAC tests
    test_role_management
    test_permission_management
    test_role_permission_relationships
    test_user_role_management

    # Clean up test data
    cleanup_test_data

    echo -e "\n${GREEN}🎉 All RBAC tests completed successfully!${NC}"
    echo -e "${BLUE}Summary:${NC}"
    echo "  - Role management: ✓"
    echo "  - Permission management: ✓"
    echo "  - Role-permission relationships: ✓"
    echo "  - User-role management: ✓"
    echo "  - Cleanup: ✓"
}

# Run main function
main "$@"
