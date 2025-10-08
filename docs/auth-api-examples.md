# 🔐 Service Authentication Examples

This document provides practical examples of authenticating with the API gateway and making authenticated calls to microservices using JWT tokens.

## 📢 Phase 2 Updates

**Important Changes in Phase 2:**

- User creation now goes through the auth service registration endpoint (`POST /api/v1/auth/register`)
- Users are no longer created directly through the user service (`POST /api/v1/users`)
- Password hashing and validation is now handled by the auth service
- Auth service communicates with user service for user data management
- All user operations (CRUD) are now properly separated between auth and user services

## 📋 Available Endpoints

### Authentication Endpoints (Public)

- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout user

### Protected Endpoints (Require Authentication)

- `GET /api/v1/auth/me` - Get current user info
- `GET /api/v1/users` - List all users
- `GET /api/v1/users/{id}` - Get user by ID
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user

**Note**: Users are created through the auth service registration endpoint (`POST /api/v1/auth/register`), not directly through the user service.

## 🛠️ Bash/Curl Examples

### Complete Authentication Flow

```bash
#!/bin/bash

# Configuration
API_BASE="http://localhost:8080"

echo "🚀 Starting authentication flow..."

# 1. Register a new user
echo "📝 Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST $API_BASE/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "password123",
    "first_name": "Demo",
    "last_name": "User"
  }')

echo "Registration Response: $REGISTER_RESPONSE"
echo

# 2. Login to get tokens
echo "🔑 Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST $API_BASE/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "password123"
  }')

# Extract tokens from response
ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.access_token')
REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.refresh_token')

echo "Access Token: ${ACCESS_TOKEN:0:50}..."
echo "Refresh Token: ${REFRESH_TOKEN:0:50}..."
echo

# 3. Get current user info (protected endpoint)
echo "👤 Getting current user info..."
USER_INFO=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
  $API_BASE/api/v1/auth/me)

echo "User Info: $USER_INFO"
echo

# 4. List all users (protected endpoint)
echo "👥 Listing all users..."
USERS_LIST=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
  $API_BASE/api/v1/users)

echo "Users List: $USERS_LIST"
echo

# 5. Create a new user through auth service registration (public endpoint)
echo "➕ Creating a new user through registration..."
CREATE_USER_RESPONSE=$(curl -s -X POST $API_BASE/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "newpassword123",
    "first_name": "New",
    "last_name": "User"
  }')

echo "Create User Response: $CREATE_USER_RESPONSE"

# Extract the new user ID (assuming response contains user data)
NEW_USER_ID=$(echo $CREATE_USER_RESPONSE | jq -r '.user.id')
echo "New User ID: $NEW_USER_ID"
echo

# 6. Get specific user by ID (protected endpoint)
if [ "$NEW_USER_ID" != "null" ] && [ -n "$NEW_USER_ID" ]; then
    echo "🔍 Getting user by ID: $NEW_USER_ID..."
    USER_BY_ID=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
      $API_BASE/api/v1/users/$NEW_USER_ID)

    echo "User by ID: $USER_BY_ID"
    echo
fi

# 7. Update user (protected endpoint)
if [ "$NEW_USER_ID" != "null" ] && [ -n "$NEW_USER_ID" ]; then
    echo "✏️ Updating user..."
    UPDATE_RESPONSE=$(curl -s -X PUT $API_BASE/api/v1/users/$NEW_USER_ID \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "first_name": "Updated",
        "last_name": "User Name"
      }')

    echo "Update Response: $UPDATE_RESPONSE"
    echo
fi

# 8. Logout (protected endpoint)
echo "🚪 Logging out..."
LOGOUT_RESPONSE=$(curl -s -X POST $API_BASE/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Logout Response: $LOGOUT_RESPONSE"
echo

echo "✅ Authentication flow completed!"
```

### Token Refresh Example

```bash
#!/bin/bash

# Function to refresh token and retry request
make_authenticated_request() {
    local method=$1
    local url=$2
    local data=$3

    # Try request with current token
    response=$(curl -s -w "%{http_code}" \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -X $method \
      -H "Content-Type: application/json" \
      ${data:+-d "$data"} \
      $url)

    http_code=${response: -3}
    response_body=${response::-3}

    # If 401 Unauthorized, try to refresh token
    if [ "$http_code" = "401" ]; then
        echo "🔄 Token expired, refreshing..."

        refresh_response=$(curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
          -H "Content-Type: application/json" \
          -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}")

        if [ $? -eq 0 ]; then
            # Update tokens
            ACCESS_TOKEN=$(echo $refresh_response | jq -r '.access_token')
            REFRESH_TOKEN=$(echo $refresh_response | jq -r '.refresh_token')

            echo "✅ Token refreshed successfully"

            # Retry the original request
            response=$(curl -s \
              -H "Authorization: Bearer $ACCESS_TOKEN" \
              -X $method \
              -H "Content-Type: application/json" \
              ${data:+-d "$data"} \
              $url)
        else
            echo "❌ Failed to refresh token"
            return 1
        fi
    fi

    echo $response
}

# Usage example
ACCESS_TOKEN="your-access-token-here"
REFRESH_TOKEN="your-refresh-token-here"

# This will automatically refresh token if expired
result=$(make_authenticated_request "GET" "http://localhost:8080/api/v1/users")
echo "Result: $result"
```

## 🐍 Python Examples

### Complete Authentication Client

```python
import requests
import json
from typing import Optional, Dict, Any

class AuthenticatedAPIClient:
    """Client for making authenticated API calls with automatic token refresh"""

    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.access_token: Optional[str] = None
        self.refresh_token: Optional[str] = None
        self.session = requests.Session()

    def register(self, email: str, password: str, first_name: str, last_name: str) -> Dict[str, Any]:
        """Register a new user"""
        data = {
            "email": email,
            "password": password,
            "first_name": first_name,
            "last_name": last_name
        }
        response = self.session.post(f"{self.base_url}/api/v1/auth/register", json=data)
        response.raise_for_status()
        return response.json()

    def login(self, email: str, password: str) -> Dict[str, Any]:
        """Login and store tokens"""
        data = {"email": email, "password": password}
        response = self.session.post(f"{self.base_url}/api/v1/auth/login", json=data)
        response.raise_for_status()

        tokens = response.json()
        self.access_token = tokens.get("access_token")
        self.refresh_token = tokens.get("refresh_token")

        # Set Authorization header for future requests
        self.session.headers.update({
            "Authorization": f"Bearer {self.access_token}"
        })

        return tokens

    def refresh_access_token(self) -> Dict[str, Any]:
        """Refresh the access token"""
        if not self.refresh_token:
            raise ValueError("No refresh token available")

        data = {"refresh_token": self.refresh_token}
        response = self.session.post(f"{self.base_url}/api/v1/auth/refresh", json=data)
        response.raise_for_status()

        tokens = response.json()
        self.access_token = tokens.get("access_token")
        self.refresh_token = tokens.get("refresh_token")

        # Update Authorization header
        self.session.headers.update({
            "Authorization": f"Bearer {self.access_token}"
        })

        return tokens

    def logout(self) -> Dict[str, Any]:
        """Logout and invalidate tokens"""
        response = self.session.post(f"{self.base_url}/api/v1/auth/logout")
        response.raise_for_status()

        self.access_token = None
        self.refresh_token = None
        self.session.headers.pop("Authorization", None)

        return response.json()

    def _make_request(self, method: str, endpoint: str, **kwargs) -> requests.Response:
        """Make authenticated request with automatic token refresh on 401"""
        url = f"{self.base_url}{endpoint}"
        response = self.session.request(method, url, **kwargs)

        # If token expired, try to refresh and retry once
        if response.status_code == 401 and self.refresh_token:
            try:
                self.refresh_access_token()
                response = self.session.request(method, url, **kwargs)
            except Exception:
                pass  # If refresh fails, return original 401 response

        return response

    # User service methods
    def get_current_user(self) -> Dict[str, Any]:
        """Get current user information"""
        response = self._make_request("GET", "/api/v1/auth/me")
        response.raise_for_status()
        return response.json()

    def get_users(self) -> list:
        """Get list of all users"""
        response = self._make_request("GET", "/api/v1/users")
        response.raise_for_status()
        return response.json()

    def create_user(self, user_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create a new user through auth service registration"""
        # Note: Users are created through registration, not directly through user service
        response = self.session.post(f"{self.base_url}/api/v1/auth/register", json=user_data)
        response.raise_for_status()
        return response.json()

    def get_user(self, user_id: str) -> Dict[str, Any]:
        """Get user by ID"""
        response = self._make_request("GET", f"/api/v1/users/{user_id}")
        response.raise_for_status()
        return response.json()

    def update_user(self, user_id: str, user_data: Dict[str, Any]) -> Dict[str, Any]:
        """Update user by ID"""
        response = self._make_request("PUT", f"/api/v1/users/{user_id}", json=user_data)
        response.raise_for_status()
        return response.json()

    def delete_user(self, user_id: str) -> Dict[str, Any]:
        """Delete user by ID"""
        response = self._make_request("DELETE", f"/api/v1/users/{user_id}")
        response.raise_for_status()
        return response.json()


# Usage example
def main():
    client = AuthenticatedAPIClient()

    try:
        # Register a new user
        print("📝 Registering user...")
        register_result = client.register(
            "python@example.com",
            "password123",
            "Python",
            "Client"
        )
        print(f"Registration successful: {register_result}")

        # Login
        print("🔑 Logging in...")
        login_result = client.login("python@example.com", "password123")
        print(f"Login successful, got tokens")

        # Get current user
        print("👤 Getting current user...")
        user_info = client.get_current_user()
        print(f"Current user: {user_info}")

        # List users
        print("👥 Listing users...")
        users = client.get_users()
        print(f"Users count: {len(users)}")

        # Create a new user through registration
        print("➕ Creating new user through registration...")
        new_user = client.create_user({
            "email": "created@example.com",
            "password": "created123",
            "first_name": "Created",
            "last_name": "User"
        })
        print(f"Created user: {new_user}")

        user_id = new_user.get("user", {}).get("id")
        if user_id:
            # Get specific user
            print(f"🔍 Getting user {user_id}...")
            user_details = client.get_user(user_id)
            print(f"User details: {user_details}")

            # Update user
            print(f"✏️ Updating user {user_id}...")
            updated_user = client.update_user(user_id, {
                "first_name": "Updated",
                "last_name": "Name"
            })
            print(f"Updated user: {updated_user}")

            # Delete user
            print(f"🗑️ Deleting user {user_id}...")
            delete_result = client.delete_user(user_id)
            print(f"Delete result: {delete_result}")

        # Logout
        print("🚪 Logging out...")
        logout_result = client.logout()
        print(f"Logout successful: {logout_result}")

    except requests.exceptions.RequestException as e:
        print(f"❌ API request failed: {e}")
    except Exception as e:
        print(f"❌ Unexpected error: {e}")


if __name__ == "__main__":
    main()
```

### Simple Token Management Example

```python
import requests
import time
from typing import Dict, Any

class SimpleAuthClient:
    """Simple client for managing authentication tokens"""

    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.access_token: str = ""
        self.refresh_token: str = ""

    def login(self, email: str, password: str) -> bool:
        """Login and get tokens"""
        try:
            response = requests.post(
                f"{self.base_url}/api/v1/auth/login",
                json={"email": email, "password": password}
            )
            response.raise_for_status()

            data = response.json()
            self.access_token = data["access_token"]
            self.refresh_token = data["refresh_token"]
            return True
        except Exception as e:
            print(f"Login failed: {e}")
            return False

    def refresh_token(self) -> bool:
        """Refresh access token"""
        try:
            response = requests.post(
                f"{self.base_url}/api/v1/auth/refresh",
                json={"refresh_token": self.refresh_token}
            )
            response.raise_for_status()

            data = response.json()
            self.access_token = data["access_token"]
            self.refresh_token = data["refresh_token"]
            return True
        except Exception as e:
            print(f"Token refresh failed: {e}")
            return False

    def make_auth_request(self, method: str, endpoint: str, **kwargs) -> requests.Response:
        """Make authenticated request with automatic token refresh"""
        headers = kwargs.get("headers", {})
        headers["Authorization"] = f"Bearer {self.access_token}"
        kwargs["headers"] = headers

        response = requests.request(method, f"{self.base_url}{endpoint}", **kwargs)

        # Auto-refresh on 401
        if response.status_code == 401:
            if self.refresh_token():
                headers["Authorization"] = f"Bearer {self.access_token}"
                response = requests.request(method, f"{self.base_url}{endpoint}", **kwargs)

        return response

# Usage
client = SimpleAuthClient()

if client.login("user@example.com", "password"):
    # Make authenticated requests
    response = client.make_auth_request("GET", "/api/v1/users")
    print(f"Users: {response.json()}")

    response = client.make_auth_request("GET", "/api/v1/auth/me")
    print(f"Current user: {response.json()}")
```

## 🔧 Testing the Examples

### Prerequisites

1. Start the services: `make dev`
2. Ensure all services are running and healthy

### Running the Bash Example

```bash
chmod +x auth-examples.sh
./auth-examples.sh
```

### Running the Python Example

```bash
pip install requests
python auth-client.py
```

## 📊 Response Examples

### Successful Login Response

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

### User Info Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "demo@example.com",
  "first_name": "Demo",
  "last_name": "User",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Error Response (401 Unauthorized)

```json
{
  "error": "Invalid or expired token",
  "type": "authentication_error"
}
```

## 🔍 Troubleshooting

### Common Issues

1. **401 Unauthorized**: Token expired or invalid

   - Solution: Refresh token or login again

2. **403 Forbidden**: Insufficient permissions

   - Solution: Check user roles and permissions

3. **Connection refused**: Services not running

   - Solution: Start services with `make dev`

4. **Invalid JSON**: Malformed request body
   - Solution: Validate JSON syntax

### Debug Tips

- Check service logs: `docker-compose logs auth-service`
- Verify tokens: Use JWT debugger online
- Test endpoints individually with curl
- Check API gateway routing configuration

## 🛡️ Role-Based Access Control (RBAC) Examples

### Admin User Registration and Authentication

```bash
#!/bin/bash

# Configuration
API_BASE="http://localhost:8080"

echo "🔐 Admin user authentication flow..."

# 1. Register admin user (requires manual database setup or admin creation endpoint)
echo "📝 Note: Admin users should be created through secure admin processes"
echo "For demo purposes, assuming admin user exists: admin@example.com"

# 2. Login as admin
echo "🔑 Logging in as admin..."
ADMIN_LOGIN=$(curl -s -X POST $API_BASE/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123"
  }')

ADMIN_ACCESS_TOKEN=$(echo $ADMIN_LOGIN | jq -r '.access_token')
ADMIN_REFRESH_TOKEN=$(echo $ADMIN_LOGIN | jq -r '.refresh_token')

echo "Admin Access Token: ${ADMIN_ACCESS_TOKEN:0:50}..."

# 3. Access admin-only endpoint (JWT key rotation)
echo "🔄 Performing admin operation: JWT key rotation..."
ROTATE_RESPONSE=$(curl -s -X POST $API_BASE/api/v1/admin/rotate-keys \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  -H "Content-Type: application/json")

echo "Rotation Response: $ROTATE_RESPONSE"

# 4. Try accessing admin endpoint with regular user token
echo "❌ Testing access denial for regular user..."
USER_LOGIN=$(curl -s -X POST $API_BASE/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "password123"
  }')

USER_ACCESS_TOKEN=$(echo $USER_LOGIN | jq -r '.access_token')

DENIED_RESPONSE=$(curl -s -X POST $API_BASE/api/v1/admin/rotate-keys \
  -H "Authorization: Bearer $USER_ACCESS_TOKEN" \
  -H "Content-Type: application/json")

echo "Access Denied Response: $DENIED_RESPONSE"
```

### Role-Based API Testing

```bash
#!/bin/bash

# Test different role-based endpoints
API_BASE="http://localhost:8080"

# Get admin token
ADMIN_TOKEN=$(curl -s -X POST $API_BASE/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}' | jq -r '.access_token')

# Get user token
USER_TOKEN=$(curl -s -X POST $API_BASE/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"password123"}' | jq -r '.access_token')

echo "🧪 Testing role-based access control..."

# Test 1: Admin can access admin endpoints
echo "✅ Admin accessing admin endpoint:"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  $API_BASE/api/v1/auth/me | jq '.email'

# Test 2: User cannot access admin endpoints
echo "❌ User trying to access admin endpoint:"
curl -s -H "Authorization: Bearer $USER_TOKEN" \
  $API_BASE/api/v1/admin/rotate-keys

# Test 3: Both can access user endpoints
echo "✅ Both admin and user can access user endpoints:"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  $API_BASE/api/v1/auth/me | jq '.roles'

curl -s -H "Authorization: Bearer $USER_TOKEN" \
  $API_BASE/api/v1/auth/me | jq '.roles'
```

## 🔗 Service-to-Service Authentication Examples

### Internal Service Communication

When services need to communicate internally (bypassing the API gateway), they should use service-specific authentication patterns:

```bash
#!/bin/bash

# Example: Auth service calling User service internally
# This would typically be done in Go code, but here's the curl equivalent

AUTH_SERVICE="http://localhost:8083"
USER_SERVICE="http://localhost:8081"

# 1. Auth service validates user and gets internal token
echo "🔐 Internal service authentication..."

# In actual implementation, services would use:
# - Shared secrets for service-to-service auth
# - Internal JWT tokens with service roles
# - mTLS certificates
# - API keys

# Example internal call (simplified)
INTERNAL_RESPONSE=$(curl -s -H "X-Service-Auth: internal-service-secret" \
  -H "X-Requested-By: auth-service" \
  $USER_SERVICE/api/v1/internal/users/validate)

echo "Internal service response: $INTERNAL_RESPONSE"
```

### Service Token Generation

```go
// In auth service - generate service-to-service tokens
func (s *AuthService) GenerateServiceToken(serviceName string) (string, error) {
    claims := &middleware.JWTClaims{
        UserID:    uuid.Nil,  // No specific user
        Email:     serviceName + "@internal",
        Roles:     []string{"service", serviceName},
        TokenType: "service",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Subject:   serviceName,
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(s.jwtUtils.GetPrivateKey())
}
```

### Service Authentication Middleware

```go
// Middleware for service-to-service authentication
func ServiceAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check for service authentication headers
        serviceAuth := c.GetHeader("X-Service-Auth")
        requestedBy := c.GetHeader("X-Requested-By")

        if serviceAuth == "" || requestedBy == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Service authentication required"})
            c.Abort()
            return
        }

        // Validate service credentials
        if !validateServiceCredentials(requestedBy, serviceAuth) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid service credentials"})
            c.Abort()
            return
        }

        c.Set("service_caller", requestedBy)
        c.Next()
    }
}
```

### Error Handling for Service Calls

```go
// Service-to-service call with error handling
func callUserService(ctx context.Context, userID string) (*User, error) {
    client := &http.Client{Timeout: 5 * time.Second}

    req, err := http.NewRequestWithContext(ctx, "GET",
        "http://user-service:8081/api/v1/internal/users/"+userID, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    // Add service authentication
    req.Header.Set("X-Service-Auth", "auth-service-secret")
    req.Header.Set("X-Requested-By", "auth-service")
    req.Header.Set("Content-Type", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("service call failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("service returned %d: %s", resp.StatusCode, string(body))
    }

    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return &user, nil
}
```

## 🔐 Advanced Authentication Patterns

### Multi-Role User Management

```bash
#!/bin/bash

# Example: User with multiple roles
API_BASE="http://localhost:8080"

# Login user with multiple roles (admin + user)
MULTI_ROLE_LOGIN=$(curl -s -X POST $API_BASE/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "multi-role@example.com",
    "password": "password123"
  }')

TOKEN=$(echo $MULTI_ROLE_LOGIN | jq -r '.access_token')

# Check user roles
echo "👤 User roles:"
curl -s -H "Authorization: Bearer $TOKEN" \
  $API_BASE/api/v1/auth/me | jq '.roles'

# Access both user and admin endpoints
echo "✅ Accessing user endpoint:"
curl -s -H "Authorization: Bearer $TOKEN" \
  $API_BASE/api/v1/users | jq '.[0].email'

echo "✅ Accessing admin endpoint:"
curl -s -H "Authorization: Bearer $TOKEN" \
  $API_BASE/api/v1/admin/rotate-keys
```

### Token Introspection

```bash
#!/bin/bash

# Check token validity and extract claims
TOKEN="your-jwt-token-here"

# Decode token payload (without verification)
PAYLOAD=$(echo $TOKEN | cut -d'.' -f2 | base64 -d 2>/dev/null | jq .)

echo "🔍 Token claims:"
echo $PAYLOAD | jq .

# Check token expiration
EXP=$(echo $PAYLOAD | jq '.exp')
NOW=$(date +%s)

if [ $EXP -lt $NOW ]; then
    echo "❌ Token expired"
else
    echo "✅ Token valid"
fi

# Extract roles
ROLES=$(echo $PAYLOAD | jq -r '.roles[]')
echo "👤 User roles: $ROLES"
```

## 📊 Response Examples with RBAC

### Successful Admin Operation

```json
{
  "message": "JWT keys rotated successfully",
  "timestamp": "2025-01-15T10:30:00Z",
  "admin_user": "admin@example.com",
  "operation": "key_rotation"
}
```

### Access Denied Response

```json
{
  "error": "Insufficient permissions",
  "required_role": "admin",
  "user_roles": ["user"],
  "endpoint": "/api/v1/admin/rotate-keys",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

### Service Authentication Error

```json
{
  "error": "Service authentication required",
  "type": "service_auth_error",
  "missing_headers": ["X-Service-Auth", "X-Requested-By"],
  "timestamp": "2025-01-15T10:30:00Z"
}
```

## 🔧 Testing Authentication & RBAC

### Automated Testing Script

```bash
#!/bin/bash

# Comprehensive auth and RBAC testing
API_BASE="http://localhost:8080"

echo "🧪 Running authentication and RBAC tests..."

# Test 1: Public endpoints (no auth required)
echo "✅ Testing public endpoints..."
curl -s -f $API_BASE/health > /dev/null && echo "Health check: PASS" || echo "Health check: FAIL"

# Test 2: Authentication flow
echo "✅ Testing authentication..."
LOGIN_RESPONSE=$(curl -s -X POST $API_BASE/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}')

if echo $LOGIN_RESPONSE | jq -e '.access_token' > /dev/null; then
    echo "Login: PASS"
    TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.access_token')
else
    echo "Login: FAIL"
    exit 1
fi

# Test 3: Protected endpoints
echo "✅ Testing protected endpoints..."
curl -s -f -H "Authorization: Bearer $TOKEN" \
  $API_BASE/api/v1/auth/me > /dev/null && echo "Protected access: PASS" || echo "Protected access: FAIL"

# Test 4: RBAC - user cannot access admin
echo "✅ Testing RBAC (user denied admin access)..."
if curl -s -H "Authorization: Bearer $TOKEN" \
  $API_BASE/api/v1/admin/rotate-keys 2>&1 | grep -q "403"; then
    echo "RBAC enforcement: PASS"
else
    echo "RBAC enforcement: FAIL"
fi

echo "🎉 All tests completed!"
```

### Load Testing Authentication

```bash
#!/bin/bash

# Load test authentication endpoints
API_BASE="http://localhost:8080"
CONCURRENT_USERS=10
REQUESTS_PER_USER=50

echo "🔥 Load testing authentication..."

# Function to simulate user authentication
simulate_user() {
    local user_id=$1

    for i in $(seq 1 $REQUESTS_PER_USER); do
        # Register user
        curl -s -X POST $API_BASE/api/v1/auth/register \
          -H "Content-Type: application/json" \
          -d "{\"email\":\"loadtest$user_id-$i@example.com\",\"password\":\"password123\",\"first_name\":\"Load\",\"last_name\":\"Test\"}" > /dev/null

        # Login
        LOGIN=$(curl -s -X POST $API_BASE/api/v1/auth/login \
          -H "Content-Type: application/json" \
          -d "{\"email\":\"loadtest$user_id-$i@example.com\",\"password\":\"password123\"}")

        TOKEN=$(echo $LOGIN | jq -r '.access_token')

        # Access protected endpoint
        curl -s -H "Authorization: Bearer $TOKEN" \
          $API_BASE/api/v1/auth/me > /dev/null
    done
}

# Run concurrent users
for user in $(seq 1 $CONCURRENT_USERS); do
    simulate_user $user &
done

wait
echo "✅ Load test completed!"
```

