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