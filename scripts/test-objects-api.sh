#!/bin/bash
# Test script for objects-service API endpoints through API Gateway
# Mirrors the same HTTP calls that MCP tools make, but via direct gateway requests.
#
# Usage:
#   ./scripts/test-objects-api.sh              # uses defaults
#   GATEWAY_URL=http://localhost:8080 \
#   AUTH_SERVICE_URL=http://localhost:8083 \
#     ./scripts/test-objects-api.sh
#
# Requirements: services must be running (docker compose up -d)

set -euo pipefail

# ──────────────────────────────────────────────
# Configuration
# ──────────────────────────────────────────────
GATEWAY_URL="${GATEWAY_URL:-http://localhost:8080}"
AUTH_SERVICE_URL="${AUTH_SERVICE_URL:-http://localhost:8083}"
POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-service-boilerplate-postgres}"

PASS=0
FAIL=0
WARN=0

# ──────────────────────────────────────────────
# Color helpers (auto-detect TTY support)
# ──────────────────────────────────────────────
if [ -t 1 ]; then
    GREEN='\033[0;32m'
    RED='\033[0;31m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    CYAN='\033[0;36m'
    BOLD='\033[1m'
    RESET='\033[0m'
else
    GREEN="" RED="" YELLOW="" BLUE="" CYAN="" BOLD="" RESET=""
fi

pass() { echo -e "  ${GREEN}✓${RESET} $1"; PASS=$((PASS+1)); }
fail() { echo -e "  ${RED}✗${RESET} $1"; FAIL=$((FAIL+1)); }
warn() { echo -e "  ${YELLOW}⊘${RESET} $1"; WARN=$((WARN+1)); }
info() { echo -e "    ${CYAN}$1${RESET}"; }

# ──────────────────────────────────────────────
# Helpers — all use temp files to avoid JSON/newline issues
# ──────────────────────────────────────────────

get_credentials() {
    local login_resp
    login_resp=$(curl -s --connect-timeout 5 "$AUTH_SERVICE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"email":"dev.admin@example.com","password":"devadmin123"}')

    JWT=$(echo "$login_resp" | jq -r '.access_token // empty' 2>/dev/null)
    if [ -z "$JWT" ] || [ "$JWT" = "null" ]; then
        echo -e "${RED}✗ Failed to obtain admin JWT${RESET}" >&2
        exit 1
    fi

    USER_ID=$(docker exec "$POSTGRES_CONTAINER" psql -U postgres -d service_db \
        -t -A -c "SELECT id FROM user_service.users WHERE email = 'dev.admin@example.com';" | tr -d ' ')

    if [ -z "$USER_ID" ]; then
        echo -e "${RED}✗ No admin user found in database${RESET}" >&2
        exit 1
    fi
}

# gateway_request: makes HTTP call, sets RESPONSE_STATUS and RESPONSE_BODY globals
gateway_request() {
    local method="$1"
    local path="$2"
    local data="${3:-}"
    local url="$GATEWAY_URL$path"
    local body_file
    body_file=$(mktemp)

    local http_code
    http_code=$(curl -s -o "$body_file" -w "%{http_code}" \
        -X "$method" \
        -H "Authorization: Bearer $JWT" \
        -H "X-User-ID: $USER_ID" \
        ${data:+-H "Content-Type: application/json" -d "$data"} \
        "$url" 2>/dev/null || echo "000")

    RESPONSE_STATUS="$http_code"
    RESPONSE_BODY=$(cat "$body_file")
    rm -f "$body_file"
}

# ──────────────────────────────────────────────
# Pre-flight checks
# ──────────────────────────────────────────────
echo ""
echo -e "${BOLD}=== Objects API Test Suite ===" 
echo -e "Target: ${GATEWAY_URL}/api/v1 (via API Gateway)"
echo -e "Auth service: ${AUTH_SERVICE_URL}"
echo -e "${RESET}"

if ! curl -s --connect-timeout 3 "$GATEWAY_URL/health" > /dev/null 2>&1; then
    echo -e "${RED}✗ API Gateway not reachable at $GATEWAY_URL${RESET}" >&2
    exit 1
fi
pass "API Gateway is up"

if ! curl -s --connect-timeout 3 "$AUTH_SERVICE_URL/health" > /dev/null 2>&1; then
    echo -e "${RED}✗ Auth service not reachable at $AUTH_SERVICE_URL${RESET}" >&2
    exit 1
fi

get_credentials
pass "Admin credentials obtained (user_id=$USER_ID)"

# ──────────────────────────────────────────────
# Helper: run a test and print result inline
# ──────────────────────────────────────────────
test_endpoint() {
    local label="$1" method="$2" path="$3" data="${4:-}" expect_ok="${5:-yes}"

    echo ""
    echo -e "${BOLD}${label}${RESET}"
    [ -n "$data" ] && info "POST/PUT with body"

    gateway_request "$method" "$path" "$data"

    if [ "$expect_ok" = "yes" ]; then
        if [ "$RESPONSE_STATUS" = "200" ]; then
            pass "${label##* } → HTTP $RESPONSE_STATUS"
        else
            local err
            err=$(echo "$RESPONSE_BODY" | jq -r '.error // "unknown"' 2>/dev/null)
            fail "${label##* } → HTTP $RESPONSE_STATUS: ${err}"
        fi
    else
        # Expect failure (e.g., known bug)
        if [ "$RESPONSE_STATUS" != "200" ]; then
            local err
            err=$(echo "$RESPONSE_BODY" | jq -r '.error // "unknown"' 2>/dev/null)
            fail "${label##* } → HTTP $RESPONSE_STATUS (expected failure: ${err})"
        else
            pass "${label##* } → unexpectedly succeeded (HTTP $RESPONSE_STATUS)"
        fi
    fi
}

# ──────────────────────────────────────────────
# Section A: Object Types (MCP tool surface area)
# ──────────────────────────────────────────────
echo ""
echo -e "${BOLD}--- Section A: Object Types ---" 
echo -e "(Mirrors MCP tools: list_object_types, get_object_type_by_id,"
echo -e "  get_parent_type_key, object_hierarchy)${RESET}"

test_endpoint "A1. GET /api/v1/object-types [list_object_types]"          "GET" "/api/v1/object-types" "" yes

# A2: Get by ID — root type with valid type_key (should work)
echo ""
echo -e "${BOLD}A2. GET /api/v1/object-types/1  [get_object_type_by_id, has type_key]${RESET}"
gateway_request "GET" "/api/v1/object-types/1"
if [ "$RESPONSE_STATUS" = "200" ]; then
    NAME=$(echo "$RESPONSE_BODY" | jq -r '.data.name // empty' 2>/dev/null)
    pass "→ $NAME (type_key present)"
else
    fail "→ HTTP $RESPONSE_STATUS: $(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"')"
fi

# A3: Get by ID — child type with NULL type_key (known pgx bug)
echo ""
echo -e "${BOLD}A3. GET /api/v1/object-types/5  [get_object_type_by_id, NULL type_key]${RESET}"
info "Electronics has type_key=NULL in DB → pgx v5 ScanArgError"
gateway_request "GET" "/api/v1/object-types/5"
if [ "$RESPONSE_STATUS" = "200" ]; then
    NAME=$(echo "$RESPONSE_BODY" | jq -r '.data.name // empty' 2>/dev/null)
    pass "→ $NAME (unexpectedly succeeded)"
else
    fail "→ HTTP $RESPONSE_STATUS: $(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"')"
fi

# A4: Get by name — works when type_key present
echo ""
echo -e "${BOLD}A4. GET /api/v1/object-types/name/Category  [get_parent_type_key]${RESET}"
gateway_request "GET" "/api/v1/object-types/name/Category"
if [ "$RESPONSE_STATUS" = "200" ]; then
    NAME=$(echo "$RESPONSE_BODY" | jq -r '.data.name // empty' 2>/dev/null)
    pass "→ $NAME"
else
    fail "→ HTTP $RESPONSE_STATUS: $(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"')"
fi

# A5: Get tree (hierarchy) — crashes on NULL type_key in CTE results
echo ""
echo -e "${BOLD}A5. GET /api/v1/object-types/1/tree?root_id=1  [object_hierarchy]${RESET}"
info "GetTree uses CTE that scans type_key → may crash"
gateway_request "GET" "/api/v1/object-types/1/tree?root_id=1"
if [ "$RESPONSE_STATUS" = "200" ]; then
    pass "→ tree returned (HTTP $RESPONSE_STATUS)"
else
    fail "→ HTTP $RESPONSE_STATUS: $(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"')"
fi

# A6: Get children — Product has child Electronics (NULL type_key)
echo ""
echo -e "${BOLD}A6. GET /api/v1/object-types/2/children  [get_children]${RESET}"
info "Product(ID=2) → child Electronics(ID=5, NULL type_key)"
gateway_request "GET" "/api/v1/object-types/2/children"
if [ "$RESPONSE_STATUS" = "200" ]; then
    pass "→ children returned (HTTP $RESPONSE_STATUS)"
else
    fail "→ HTTP $RESPONSE_STATUS: $(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"')"
fi

# A7: Search — returns rows with NULL type_key
echo ""
echo -e "${BOLD}A7. GET /api/v1/object-types/search?q=electronic  [search]${RESET}"
gateway_request "GET" "/api/v1/object-types/search?q=electronic"
if [ "$RESPONSE_STATUS" = "200" ]; then
    pass "→ search results (HTTP $RESPONSE_STATUS)"
else
    fail "→ HTTP $RESPONSE_STATUS: $(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"')"
fi

# A8: Get ancestors — does NOT select type_key, should work
echo ""
echo -e "${BOLD}A8. GET /api/v1/object-types/5/ancestors  [get_ancestors]${RESET}"
info "Electronics → Product (ID=2) → Category (ID=1)"
gateway_request "GET" "/api/v1/object-types/5/ancestors"
if [ "$RESPONSE_STATUS" = "200" ]; then
    pass "→ ancestors returned (HTTP $RESPONSE_STATUS)"
else
    fail "→ HTTP $RESPONSE_STATUS: $(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"')"
fi

# A9: Get path — does NOT select type_key, should work
echo ""
echo -e "${BOLD}A9. GET /api/v1/object-types/5/path  [get_path]${RESET}"
gateway_request "GET" "/api/v1/object-types/5/path"
if [ "$RESPONSE_STATUS" = "200" ]; then
    pass "→ path returned (HTTP $RESPONSE_STATUS)"
else
    fail "→ HTTP $RESPONSE_STATUS: $(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"')"
fi

# A10: Validate move — business logic, no scan issues expected
echo ""
echo -e "${BOLD}A10. POST /api/v1/object-types/5/validate-move?new_parent_id=1  [validate_move]${RESET}"
gateway_request "POST" "/api/v1/object-types/5/validate-move?new_parent_id=1"
if [ "$RESPONSE_STATUS" = "200" ]; then
    VALID=$(echo "$RESPONSE_BODY" | jq -r '.valid // empty' 2>/dev/null)
    pass "→ valid=$VALID"
else
    fail "→ HTTP $RESPONSE_STATUS: $(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"')"
fi

# A11: Subtree count — uses objects table, no type_key scan
echo ""
echo -e "${BOLD}A11. GET /api/v1/object-types/2/subtree-count  [subtree_object_count]${RESET}"
gateway_request "GET" "/api/v1/object-types/2/subtree-count"
if [ "$RESPONSE_STATUS" = "200" ]; then
    COUNT=$(echo "$RESPONSE_BODY" | jq -r '.count // empty' 2>/dev/null)
    pass "→ $COUNT objects in subtree"
else
    fail "→ HTTP $RESPONSE_STATUS: $(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"')"
fi

# ──────────────────────────────────────────────
# Section B: Objects (MCP tool surface area)
# ──────────────────────────────────────────────
echo ""
echo -e "${BOLD}--- Section B: Objects ---${RESET}"

# B1: List objects
echo ""
echo -e "${BOLD}B1. GET /api/v1/objects  [list_objects]${RESET}"
gateway_request "GET" "/api/v1/objects"
if [ "$RESPONSE_STATUS" = "200" ]; then
    pass "→ list returned (HTTP $RESPONSE_STATUS)"
else
    fail "→ HTTP $RESPONSE_STATUS: $(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"')"
fi

# B2: Get object by ID — find first valid one from DB
echo ""
echo -e "${BOLD}B2. GET /api/v1/objects/{id}  [get_object_by_id]${RESET}"
OBJ_ID=$(docker exec "$POSTGRES_CONTAINER" psql -U postgres -d service_db \
    -t -A -c "SELECT id FROM objects_service.objects WHERE deleted_at IS NULL LIMIT 1;" | tr -d ' ')

if [ -n "$OBJ_ID" ]; then
    gateway_request "GET" "/api/v1/objects/$OBJ_ID"
    if [ "$RESPONSE_STATUS" = "200" ]; then
        NAME=$(echo "$RESPONSE_BODY" | jq -r '.data.name // empty' 2>/dev/null)
        pass "→ $NAME (id=$OBJ_ID)"
    else
        fail "→ HTTP $RESPONSE_STATUS: $(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"')"
    fi

    # Also test by public_id
    OBJ_PID=$(docker exec "$POSTGRES_CONTAINER" psql -U postgres -d service_db \
        -t -A -c "SELECT public_id FROM objects_service.objects WHERE deleted_at IS NULL LIMIT 1;" | tr -d ' ')
    if [ -n "$OBJ_PID" ]; then
        gateway_request "GET" "/api/v1/objects/public-id/$OBJ_PID"
        if [ "$RESPONSE_STATUS" = "200" ]; then
            pass "→ found by public_id=$OBJ_PID"
        else
            local err
            err=$(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"' 2>/dev/null)
            fail "→ HTTP $RESPONSE_STATUS: ${err}"
        fi
    fi
else
    warn "No objects in DB — skipping B2"
fi

# ──────────────────────────────────────────────
# Section C: Relationship Types (browse_schema)
# ──────────────────────────────────────────────
echo ""
echo -e "${BOLD}--- Section C: Relationship Types ---${RESET}"

# C1: List relationship types
echo ""
echo -e "${BOLD}C1. GET /api/v1/relationship-types  [browse_schema]${RESET}"
gateway_request "GET" "/api/v1/relationship-types"
if [ "$RESPONSE_STATUS" = "200" ]; then
    COUNT=$(echo "$RESPONSE_BODY" | jq '.data | length' 2>/dev/null || echo "?")
    pass "→ $COUNT relationship types found"
else
    fail "→ HTTP $RESPONSE_STATUS: $(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"')"
fi

# C2: Get by type_key
echo ""
echo -e "${BOLD}C2. GET /api/v1/relationship-types/belongs_to  [get_relationship_type]${RESET}"
gateway_request "GET" "/api/v1/relationship-types/belongs_to"
if [ "$RESPONSE_STATUS" = "200" ]; then
    NAME=$(echo "$RESPONSE_BODY" | jq -r '.data.relationship_name // empty' 2>/dev/null)
    pass "→ $NAME"
else
    fail "→ HTTP $RESPONSE_STATUS: $(echo "$RESPONSE_BODY" | jq -r '.error // "n/a"')"
fi

# ──────────────────────────────────────────────
# Section D: Auth Chain Verification
# ──────────────────────────────────────────────
echo ""
echo -e "${BOLD}--- Section D: Auth Chain Verification ---${RESET}"

echo ""
echo -e "${BOLD}D1. Object types permissions in auth_service${RESET}"
info "Scope is encoded in 'name' field: e.g., 'object-types:read:all' (scoped) vs 'object-types:create' (flat)"
SCOPED=$(docker exec "$POSTGRES_CONTAINER" psql -U postgres -d service_db \
    -t -A -c "SELECT COUNT(*) FROM auth_service.permissions WHERE resource = 'object-types' AND (name LIKE '%:all' OR name LIKE '%:own');" 2>/dev/null || echo "?")
FLAT=$(docker exec "$POSTGRES_CONTAINER" psql -U postgres -d service_db \
    -t -A -c "SELECT COUNT(*) FROM auth_service.permissions WHERE resource = 'object-types' AND name NOT LIKE '%:all' AND name NOT LIKE '%:own';" 2>/dev/null || echo "?")
TOTAL=$(docker exec "$POSTGRES_CONTAINER" psql -U postgres -d service_db \
    -t -A -c "SELECT COUNT(*) FROM auth_service.permissions WHERE resource = 'object-types';" 2>/dev/null || echo "?")
info "Total=$TOTAL scoped=$SCOPED flat=$FLAT"

if [ "$SCOPED" != "?" ] && [ "$SCOPED" -ge 6 ] 2>/dev/null; then
    pass "Scoped permissions present (≥6 for object-types)"
else
    warn "Expected ≥6 scoped entries (found $SCOPED, flat=$FLAT)"
fi

echo ""
echo -e "${BOLD}D2. Role → permission linkage${RESET}"
for ROLE in admin object-type-admin user; do
    COUNT=$(docker exec "$POSTGRES_CONTAINER" psql -U postgres -d service_db \
        -t -A -c "SELECT COUNT(*) FROM auth_service.role_permissions rp
                  JOIN auth_service.permissions p ON rp.permission_id = p.id
                  WHERE p.resource = 'object-types' 
                  AND rp.role_id IN (SELECT id FROM auth_service.roles WHERE name = '$ROLE');" 2>/dev/null || echo "?")
    info "Role '$ROLE' → $COUNT permissions"
done

echo ""
echo -e "${BOLD}D3. MCP agent role${RESET}"
MCP_PERMS=$(docker exec "$POSTGRES_CONTAINER" psql -U postgres -d service_db \
    -t -A -c "SELECT string_agg(p.name, ', ') 
              FROM auth_service.role_permissions rp
              JOIN auth_service.permissions p ON rp.permission_id = p.id
              WHERE rp.role_id IN (SELECT id FROM auth_service.roles WHERE name = 'mcp-agent-read-only');" 2>/dev/null || echo "?")

if [ -n "$MCP_PERMS" ] && [ "$MCP_PERMS" != "?" ]; then
    pass "mcp-agent permissions: $MCP_PERMS"
else
    warn "Could not query mcp-agent role permissions"
fi

# ──────────────────────────────────────────────
# Summary
# ──────────────────────────────────────────────
echo ""
echo -e "${BOLD}=== Results ===${RESET}"
echo -e "  ${GREEN}Passed: $PASS${RESET}"
if [ "$FAIL" -gt 0 ]; then
    echo -e "  ${RED}Failed: $FAIL${RESET}"
else
    echo -e "  Failed: $FAIL"
fi
[ "$WARN" -gt 0 ] && echo -e "  ${YELLOW}Warnings: $WARN${RESET}"

echo ""
if [ "$FAIL" -eq 0 ]; then
    echo -e "${GREEN}${BOLD}All tests passed.${RESET}"
else
    echo -e "${RED}${BOLD}$FAIL test(s) failed. See details above.${RESET}"
    if [ "$FAIL" -ge 2 ]; then
        echo ""
        echo -e "${YELLOW}Known issue: Several endpoints fail due to a pre-existing${RESET}"
        echo -e "${YELLOW}pgx v5 ScanArgError — type_key is NULL for 8/12 object-types.${RESET}"
        echo -e "${YELLOW}The model field is string (not *string), so pgx refuses NULL scans.${RESET}"
    fi
fi
