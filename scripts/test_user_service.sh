#!/bin/bash

# Comprehensive User Service API Testing Script
# Tests all CRUD operations and error handling scenarios

BASE_URL="http://localhost:8081/api/v1"
USERS_ENDPOINT="$BASE_URL/users"

echo "=== User Service API Testing ==="
echo "Base URL: $BASE_URL"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "SUCCESS")
            echo -e "${GREEN}✓ $message${NC}"
            ;;
        "ERROR")
            echo -e "${RED}✗ $message${NC}"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ $message${NC}"
            ;;
        "WARNING")
            echo -e "${YELLOW}⚠ $message${NC}"
            ;;
    esac
}

# Function to make curl request and capture response
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local description=$4

    echo
    print_status "INFO" "Testing: $description"
    echo "Method: $method"
    echo "URL: $url"
    if [ -n "$data" ]; then
        echo "Data: $data"
    fi
    echo "Response:"

    if [ "$method" = "GET" ] || [ "$method" = "DELETE" ]; then
        response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X $method "$url" 2>/dev/null)
    else
        response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X $method -H "Content-Type: application/json" -d "$data" "$url" 2>/dev/null)
    fi

    # Extract status code and body
    body=$(echo "$response" | sed '$d')
    status_code=$(echo "$response" | grep "HTTP_STATUS:" | cut -d: -f2)

    echo "$body" | jq . 2>/dev/null || echo "$body"
    echo "Status Code: $status_code"

    # Return status code for caller
    return $status_code
}

echo "=== Testing User Service Health ==="
make_request "GET" "$BASE_URL/ping" "" "Health check"
echo

echo "=== CREATE USER TESTS ==="

# Test 1: Create valid user
echo "Test 1: Create valid user"
make_request "POST" "$USERS_ENDPOINT" '{
    "email": "testuser1@example.com",
    "first_name": "Test",
    "last_name": "User",
    "password": "password123"
}' "Create valid user"
user1_id=$(echo "$body" | jq -r '.data.id' 2>/dev/null || echo "")

# Test 2: Create duplicate user (should fail)
echo "Test 2: Create duplicate user (should fail with 409)"
make_request "POST" "$USERS_ENDPOINT" '{
    "email": "testuser1@example.com",
    "first_name": "Test",
    "last_name": "User",
    "password": "password123"
}' "Create duplicate user"

# Test 3: Create user with invalid email (should fail)
echo "Test 3: Create user with invalid email (should fail with 400)"
make_request "POST" "$USERS_ENDPOINT" '{
    "email": "invalid-email",
    "first_name": "Test",
    "last_name": "User",
    "password": "password123"
}' "Create user with invalid email"

# Test 4: Create user with missing required fields (should fail)
echo "Test 4: Create user with missing email (should fail with 400)"
make_request "POST" "$USERS_ENDPOINT" '{
    "first_name": "Test",
    "last_name": "User",
    "password": "password123"
}' "Create user with missing email"

# Test 5: Create another valid user for further testing
echo "Test 5: Create another valid user"
make_request "POST" "$USERS_ENDPOINT" '{
    "email": "testuser2@example.com",
    "first_name": "Jane",
    "last_name": "Doe",
    "password": "password456"
}' "Create another valid user"
user2_id=$(echo "$body" | jq -r '.data.id' 2>/dev/null || echo "")

echo
echo "=== GET USER TESTS ==="

# Test 6: Get existing user
echo "Test 6: Get existing user"
if [ -n "$user1_id" ]; then
    make_request "GET" "$USERS_ENDPOINT/$user1_id" "" "Get existing user"
else
    print_status "WARNING" "Skipping test - no user1_id available"
fi

# Test 7: Get non-existing user (should fail with 404)
echo "Test 7: Get non-existing user (should fail with 404)"
make_request "GET" "$USERS_ENDPOINT/99999" "" "Get non-existing user"

# Test 8: Get user with invalid ID format (should fail with 400)
echo "Test 8: Get user with invalid ID format (should fail with 400)"
make_request "GET" "$USERS_ENDPOINT/abc" "" "Get user with invalid ID"

echo
echo "=== UPDATE USER TESTS ==="

# Test 9: Update existing user
echo "Test 9: Update existing user"
if [ -n "$user1_id" ]; then
    make_request "PUT" "$USERS_ENDPOINT/$user1_id" '{
        "first_name": "Updated",
        "last_name": "Name"
    }' "Update existing user"
else
    print_status "WARNING" "Skipping test - no user1_id available"
fi

# Test 10: Update non-existing user (should fail with 404)
echo "Test 10: Update non-existing user (should fail with 404)"
make_request "PUT" "$USERS_ENDPOINT/99999" '{
    "first_name": "Updated",
    "last_name": "Name"
}' "Update non-existing user"

# Test 11: Update user with invalid data (should fail with 400)
echo "Test 11: Update user with invalid email (should fail with 400)"
if [ -n "$user1_id" ]; then
    make_request "PUT" "$USERS_ENDPOINT/$user1_id" '{
        "email": "invalid-email-format"
    }' "Update user with invalid email"
else
    print_status "WARNING" "Skipping test - no user1_id available"
fi

echo
echo "=== DELETE USER TESTS ==="

# Test 12: Delete existing user
echo "Test 12: Delete existing user"
if [ -n "$user2_id" ]; then
    make_request "DELETE" "$USERS_ENDPOINT/$user2_id" "" "Delete existing user"
else
    print_status "WARNING" "Skipping test - no user2_id available"
fi

# Test 13: Delete non-existing user (should fail with 404)
echo "Test 13: Delete non-existing user (should fail with 404)"
make_request "DELETE" "$USERS_ENDPOINT/99999" "" "Delete non-existing user"

# Test 14: Delete user with invalid ID format (should fail with 400)
echo "Test 14: Delete user with invalid ID format (should fail with 400)"
make_request "DELETE" "$USERS_ENDPOINT/abc" "" "Delete user with invalid ID"

echo
echo "=== LIST USERS TESTS ==="

# Test 15: List users with default pagination
echo "Test 15: List users with default pagination"
make_request "GET" "$USERS_ENDPOINT" "" "List users (default pagination)"

# Test 16: List users with custom pagination
echo "Test 16: List users with custom pagination (limit=5, offset=0)"
make_request "GET" "$USERS_ENDPOINT?limit=5&offset=0" "" "List users with custom pagination"

# Test 17: List users with invalid limit (should default to 10)
echo "Test 17: List users with invalid limit (should default to 10)"
make_request "GET" "$USERS_ENDPOINT?limit=abc" "" "List users with invalid limit"

# Test 18: List users with limit too high (should fail with 400)
echo "Test 18: List users with limit too high (should fail with 400)"
make_request "GET" "$USERS_ENDPOINT?limit=150" "" "List users with limit too high"

echo
echo "=== Testing Complete ==="
echo "Summary of tests performed:"
echo "- Create user (valid, duplicate, invalid email, missing fields)"
echo "- Get user (existing, non-existing, invalid ID)"
echo "- Update user (existing, non-existing, invalid data)"
echo "- Delete user (existing, non-existing, invalid ID)"
echo "- List users (default, custom pagination, invalid params)"
echo
print_status "INFO" "All tests completed. Check the responses above for proper error handling."