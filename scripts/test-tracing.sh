#!/bin/bash

# Test script to demonstrate database tracing instrumentation
# This script assumes all services are already running and makes API calls to generate traces

set -e

echo "🔍 Making test API calls to generate database traces..."
echo "   (Assumes services are already running with 'make dev')"
echo ""

# Register a test user (this will generate INSERT traces in both auth-service and user-service)
echo "📝 Registering test user via auth service..."
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test-tracing@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "Tracing"
  }' || echo "User registration failed (may already exist)"

# Login to get auth token (this will generate SELECT trace)
echo ""
echo "🔐 Logging in to get auth token..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test-tracing@example.com",
    "password": "password123"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
  echo "❌ Failed to get auth token"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo "✅ Got auth token"

# Get current user info (this will generate SELECT trace)
echo ""
echo "👤 Getting current user info..."
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN" || echo "Failed to get user info"

# List users (this will generate SELECT trace with LIMIT/OFFSET)
echo ""
echo "📋 Listing users..."
curl -X GET "http://localhost:8080/api/v1/users?limit=5&offset=0" \
  -H "Authorization: Bearer $TOKEN" || echo "Failed to list users"

echo ""
echo "✅ Test API calls completed!"
echo ""
echo "🔍 View traces in Jaeger UI: http://localhost:16686"
echo "   - Look for spans named 'db.INSERT', 'db.SELECT', 'db.TRANSACTION'"
echo "   - Check both 'auth-service' and 'user-service' traces"
echo "   - Note the timing information for database operations"
echo ""
echo "💡 Tip: Refresh the Jaeger UI to see new traces"

