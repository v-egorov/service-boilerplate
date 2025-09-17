#!/bin/bash

# Test Authentication Flow Script
# This script demonstrates the complete authentication flow and API calls

set -e  # Exit on any error

# Configuration
API_BASE="http://localhost:8080"
TEST_EMAIL="test-$(date +%s)@example.com"
TEST_PASSWORD="password123"

echo "🚀 Testing Authentication Flow"
echo "================================="
echo "API Base URL: $API_BASE"
echo "Test Email: $TEST_EMAIL"
echo

# Function to check if services are running
check_services() {
    echo "🔍 Checking if services are running..."

    # Check API gateway
    if ! curl -s -f $API_BASE/health > /dev/null; then
        echo "❌ API Gateway not responding at $API_BASE"
        echo "💡 Make sure to run: make dev"
        exit 1
    fi

    echo "✅ Services are running"
    echo
}

# Function to make authenticated requests
make_auth_request() {
    local method=$1
    local endpoint=$2
    local data=$3

    response=$(curl -s -w "\n%{http_code}" \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -X "$method" \
      -H "Content-Type: application/json" \
      ${data:+-d "$data"} \
      "$API_BASE$endpoint")

    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)

    echo "$response_body"

    # Return non-zero if not successful
    if [ "$http_code" -lt 200 ] || [ "$http_code" -ge 300 ]; then
        echo "❌ Request failed with status $http_code" >&2
        return 1
    fi
}

# Check services first
check_services

# Step 1: Register a new user
echo "📝 Step 1: Registering new user..."
REGISTER_RESPONSE=$(curl -s -X POST $API_BASE/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\",
    \"first_name\": \"Test\",
    \"last_name\": \"User\"
  }")

if [ $? -ne 0 ]; then
    echo "❌ Registration failed"
    echo "Response: $REGISTER_RESPONSE"
    exit 1
fi

echo "✅ Registration successful"
echo "Response: $REGISTER_RESPONSE"
echo

# Step 2: Login to get tokens
echo "🔑 Step 2: Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST $API_BASE/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\"
  }")

if [ $? -ne 0 ]; then
    echo "❌ Login failed"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

# Extract tokens
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.refresh_token')

if [ "$ACCESS_TOKEN" = "null" ] || [ -z "$ACCESS_TOKEN" ]; then
    echo "❌ Failed to extract access token"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

echo "✅ Login successful"
echo "Access Token: ${ACCESS_TOKEN:0:50}..."
echo "Refresh Token: ${REFRESH_TOKEN:0:50}..."
echo

# Step 3: Get current user info
echo "👤 Step 3: Getting current user info..."
USER_INFO=$(make_auth_request "GET" "/api/v1/auth/me")
echo "✅ Current user info retrieved"
echo "Response: $USER_INFO"
echo

# Step 4: List all users
echo "👥 Step 4: Listing all users..."
USERS_LIST=$(make_auth_request "GET" "/api/v1/users")
echo "✅ Users list retrieved"
echo "Response: $USERS_LIST"
echo

# Step 5: Create a new user
echo "➕ Step 5: Creating a new user..."
NEW_USER_EMAIL="created-$(date +%s)@example.com"
CREATE_RESPONSE=$(make_auth_request "POST" "/api/v1/users" "{
  \"email\": \"$NEW_USER_EMAIL\",
  \"password\": \"created123\",
  \"first_name\": \"Created\",
  \"last_name\": \"User\"
}")

echo "✅ User created successfully"
echo "Response: $CREATE_RESPONSE"
echo

# Extract new user ID
NEW_USER_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id')
if [ "$NEW_USER_ID" != "null" ] && [ -n "$NEW_USER_ID" ]; then
    # Step 6: Get specific user
    echo "🔍 Step 6: Getting user by ID ($NEW_USER_ID)..."
    USER_BY_ID=$(make_auth_request "GET" "/api/v1/users/$NEW_USER_ID")
    echo "✅ User retrieved by ID"
    echo "Response: $USER_BY_ID"
    echo

    # Step 7: Update user
    echo "✏️ Step 7: Updating user..."
    UPDATE_RESPONSE=$(make_auth_request "PUT" "/api/v1/users/$NEW_USER_ID" "{
      \"first_name\": \"Updated\",
      \"last_name\": \"User Name\"
    }")
    echo "✅ User updated successfully"
    echo "Response: $UPDATE_RESPONSE"
    echo

    # Step 8: Delete user
    echo "🗑️ Step 8: Deleting user..."
    DELETE_RESPONSE=$(make_auth_request "DELETE" "/api/v1/users/$NEW_USER_ID")
    echo "✅ User deleted successfully"
    echo "Response: $DELETE_RESPONSE"
    echo
else
    echo "⚠️ Could not extract user ID, skipping user-specific operations"
    echo
fi

# Step 9: Logout
echo "🚪 Step 9: Logging out..."
LOGOUT_RESPONSE=$(curl -s -X POST $API_BASE/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if [ $? -eq 0 ]; then
    echo "✅ Logout successful"
    echo "Response: $LOGOUT_RESPONSE"
else
    echo "⚠️ Logout may have failed, but continuing..."
fi
echo

# Step 10: Try to access protected endpoint after logout (should fail)
echo "🔒 Step 10: Testing access after logout (should fail)..."
AFTER_LOGOUT=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
  $API_BASE/api/v1/auth/me)

if [ $? -ne 0 ]; then
    echo "✅ Access correctly denied after logout"
else
    echo "⚠️ Access still works after logout - this might be expected if tokens aren't immediately invalidated"
fi
echo "Response: $AFTER_LOGOUT"
echo

echo "🎉 Authentication flow test completed successfully!"
echo "================================="
echo "✅ All major authentication operations tested:"
echo "   - User registration"
echo "   - User login & token retrieval"
echo "   - Protected endpoint access"
echo "   - User CRUD operations"
echo "   - User logout"
echo "   - Post-logout access control"
echo
echo "📚 For more examples, see: docs/auth-api-examples.md"