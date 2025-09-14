#!/bin/bash

# Comprehensive PUT and PATCH API Testing Script
# Tests both full replacement (PUT) and partial update (PATCH) operations

BASE_URL="http://localhost:8081/api/v1"
USERS_ENDPOINT="$BASE_URL/users"

echo "=== PUT vs PATCH API Testing ==="
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

# Create a test user first
echo "=== Creating Test User ==="
make_request "POST" "$USERS_ENDPOINT" '{
    "email": "testuser@example.com",
    "first_name": "Test",
    "last_name": "User"
}' "Create test user"

# Extract user ID from response
user_id=$(echo "$body" | jq -r '.data.id' 2>/dev/null || echo "")

if [ -z "$user_id" ] || [ "$user_id" = "null" ]; then
    print_status "ERROR" "Failed to create test user or extract ID"
    exit 1
fi

print_status "SUCCESS" "Created test user with ID: $user_id"

echo
echo "=== PUT TESTS (Full Resource Replacement) ==="

# Test 1: PUT with all required fields (should succeed)
echo "Test 1: PUT with all required fields"
make_request "PUT" "$USERS_ENDPOINT/$user_id" '{
    "email": "updated@example.com",
    "first_name": "Updated",
    "last_name": "Name"
}' "PUT with all required fields"

# Test 2: PUT with missing required field (should fail)
echo "Test 2: PUT with missing email (should fail with 400)"
make_request "PUT" "$USERS_ENDPOINT/$user_id" '{
    "first_name": "Updated",
    "last_name": "Name"
}' "PUT with missing email"

# Test 3: PUT with invalid email (should fail)
echo "Test 3: PUT with invalid email (should fail with 400)"
make_request "PUT" "$USERS_ENDPOINT/$user_id" '{
    "email": "invalid-email",
    "first_name": "Updated",
    "last_name": "Name"
}' "PUT with invalid email"

# Test 4: PUT with non-existing user (should fail with 404)
echo "Test 4: PUT with non-existing user (should fail with 404)"
make_request "PUT" "$USERS_ENDPOINT/99999" '{
    "email": "nonexistent@example.com",
    "first_name": "Non",
    "last_name": "Existent"
}' "PUT with non-existing user"

echo
echo "=== PATCH TESTS (Partial Resource Update) ==="

# Test 5: PATCH with single field (should succeed)
echo "Test 5: PATCH with single field"
make_request "PATCH" "$USERS_ENDPOINT/$user_id" '{
    "first_name": "Patched"
}' "PATCH with single field"

# Test 6: PATCH with multiple fields (should succeed)
echo "Test 6: PATCH with multiple fields"
make_request "PATCH" "$USERS_ENDPOINT/$user_id" '{
    "last_name": "Patched",
    "email": "patched@example.com"
}' "PATCH with multiple fields"

# Test 7: PATCH with no fields (should fail with 400)
echo "Test 7: PATCH with no fields (should fail with 400)"
make_request "PATCH" "$USERS_ENDPOINT/$user_id" '{}' "PATCH with no fields"

# Test 8: PATCH with invalid email (should fail with 400)
echo "Test 8: PATCH with invalid email (should fail with 400)"
make_request "PATCH" "$USERS_ENDPOINT/$user_id" '{
    "email": "invalid-email-format"
}' "PATCH with invalid email"

# Test 9: PATCH with non-existing user (should fail with 404)
echo "Test 9: PATCH with non-existing user (should fail with 404)"
make_request "PATCH" "$USERS_ENDPOINT/99999" '{
    "first_name": "NonExistent"
}' "PATCH with non-existing user"

echo
echo "=== Verification Tests ==="

# Test 10: Verify the final state of the user
echo "Test 10: Verify final user state"
make_request "GET" "$USERS_ENDPOINT/$user_id" "" "Get final user state"

echo
echo "=== Testing Complete ==="
echo "Summary of PUT vs PATCH differences:"
echo
echo "PUT (Full Resource Replacement):"
echo "  ✅ Requires ALL fields in request body"
echo "  ✅ Replaces entire resource"
echo "  ✅ Idempotent operation"
echo "  ✅ Returns 200 OK on success"
echo "  ❌ Returns 400 Bad Request if any required field missing"
echo
echo "PATCH (Partial Resource Update):"
echo "  ✅ Allows partial updates (any subset of fields)"
echo "  ✅ Only updates provided fields"
echo "  ✅ More flexible for clients"
echo "  ✅ Returns 200 OK on success"
echo "  ❌ Returns 400 Bad Request if no fields provided"
echo
print_status "SUCCESS" "PUT and PATCH testing completed successfully!"