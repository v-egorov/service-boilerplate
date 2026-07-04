#!/bin/bash
# MCP Server E2E Test — verifies the full protocol handshake through API Gateway
# Note: mcp-go SSE transport returns responses via the SSE stream (async),
# not as direct HTTP response bodies. This script keeps an SSE listener alive
# and reads responses from it after each request.

set -euo pipefail

BASE_URL="${MCP_BASE_URL:-http://localhost:8080}"
MCP_SERVER_URL="http://localhost:8095"
TIMEOUT=15
PASS=0
FAIL=0

# Color helpers (auto-detect TTY support)
if [ -t 1 ]; then
    GREEN='\033[0;32m'
    RED='\033[0;31m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    BOLD='\033[1m'
    RESET='\033[0m'
else
    GREEN="" RED="" YELLOW="" BLUE="" BOLD="" RESET=""
fi

pass() { echo -e "  ${GREEN}✓${RESET} $1"; PASS=$((PASS+1)); }
fail() { echo -e "  ${RED}✗${RESET} $1"; FAIL=$((FAIL+1)); }
warn() { echo -e "  ${YELLOW}⊘${RESET} $1"; }
info() { echo -e "    $BLUE$1${RESET}"; }

cleanup() { kill $SSE_PID 2>/dev/null || true; wait $SSE_PID 2>/dev/null || true; rm -f "$SSE_OUT" "$RESPONSES"; }
trap cleanup EXIT

echo ""
echo -e "${BOLD}=== MCP Server E2E Test ===" 
echo -e "Target: ${BASE_URL}/mcp (via API Gateway)"
echo ""

# ──────────────────────────────────────────────
# Step 1: Verify mcp-server is running (direct)
# ──────────────────────────────────────────────
echo -e "${BOLD}Step 1: mcp-server health check${RESET}"
HEALTH=$(curl -s --connect-timeout 5 "$MCP_SERVER_URL/health" \
    | jq -r '.data.status // empty' 2>/dev/null || echo "unreachable")

if [ "$HEALTH" = "ok" ]; then
    pass "mcp-server healthy (port 8095)"
else
    fail "mcp-server unreachable at $MCP_SERVER_URL (status: $HEALTH)"
    echo ""
    echo "Cannot continue — mcp-server must be running."
    exit 1
fi

# ──────────────────────────────────────────────
# Step 2: SSE connection → establish session
# ──────────────────────────────────────────────
echo -e "${BOLD}Step 2: Establish SSE session through Gateway${RESET}"
info "GET $BASE_URL/mcp/sse"

SSE_OUT=$(mktemp)
RESPONSES=$(mktemp)
curl -sN --connect-timeout 5 "$BASE_URL/mcp/sse" > "$SSE_OUT" &
SSE_PID=$!
sleep 2

# mcp-go sends: "data: /message?sessionId=..."
SUBMIT_PATH=$(grep "^data:" "$SSE_OUT" | head -1 | sed 's/^data: //' | tr -d '\r' || true)

if [ -n "$SUBMIT_PATH" ]; then
    SESSION_ID=$(echo "$SUBMIT_PATH" | grep -oP 'sessionId=\K[^&]+' || echo "?")
    SUBMIT_URL="${BASE_URL}/mcp/message?sessionId=${SESSION_ID}"
    info "Session ID: ${SESSION_ID:0:8}..."
    info "Submit URL: $SUBMIT_URL"
    pass "SSE connection established"
else
    fail "SSE endpoint returned no message submission URL"
    echo "  Raw SSE output:"
    cat "$SSE_OUT" | head -5 | sed 's/^/      /'
    rm -f "$SSE_OUT" "$RESPONSES"
    exit 1
fi

# ──────────────────────────────────────────────
# Step 3: MCP Initialize handshake
# ──────────────────────────────────────────────
echo ""
echo -e "${BOLD}Step 3: MCP Initialize${RESET}"
info "POST $SUBMIT_URL"

curl -s --connect-timeout 5 "$SUBMIT_URL" \
    -H "Content-Type: application/json" \
    -d '{
        "jsonrpc": "2.0",
        "id": 1,
        "method": "initialize",
        "params": {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "e2e-test-client", "version": "1.0"}
        }
    }' > /dev/null

# Wait for SSE response (mcp-go sends init response back via SSE stream)
sleep 1

# Extract JSON from SSE data lines (strip 'data: ' prefix)
extract_json() {
    grep '^data:' "$1" | grep "$2" | tail -1 | sed 's/^data: //' | tr -d '\r'
}

INIT_RESPONSE=$(extract_json "$SSE_OUT" '"id":1')

if [ -n "$INIT_RESPONSE" ]; then
    SERVER_NAME=$(echo "$INIT_RESPONSE" | jq -r '.result.serverInfo.name // empty' 2>/dev/null || echo "?")
    PROTO_VER=$(echo "$INIT_RESPONSE" | jq -r '.result.protocolVersion // empty' 2>/dev/null || echo "?")
    pass "Initialize accepted (protocol=$PROTO_VER, server=$SERVER_NAME)"
else
    fail "No initialize response received via SSE stream"
fi

# ──────────────────────────────────────────────
# Step 4: List available tools
# ──────────────────────────────────────────────
echo ""
echo -e "${BOLD}Step 4: List MCP Tools${RESET}"
info "POST $SUBMIT_URL (tools/list)"

curl -s --connect-timeout 5 "$SUBMIT_URL" \
    -H "Content-Type: application/json" \
    -d '{
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/list",
        "params": {}
    }' > /dev/null

# Wait for SSE response
sleep 1

TOOLS_RESPONSE=$(extract_json "$SSE_OUT" '"id":2')

TOOL_COUNT=$(echo "$TOOLS_RESPONSE" | jq '.result.tools | length' 2>/dev/null || echo "0")

if [ "$TOOL_COUNT" -gt 0 ] 2>/dev/null; then
    pass "Found $TOOL_COUNT MCP tool(s)"
    # Show each tool name with description snippet
    echo "$TOOLS_RESPONSE" | jq -r '.result.tools[] | "    \(.name): \(if (.description | length) > 60 then .description[:57] + "..." else .description end)"' 2>/dev/null | while read line; do
        info "$line"
    done
else
    fail "No tools returned from server"
    if [ -n "$TOOLS_RESPONSE" ]; then
        info "Response:"
        echo "$TOOLS_RESPONSE" | head -3 | sed 's/^/      /'
    fi
fi

# ──────────────────────────────────────────────
# Step 5: Call list_object_types (full chain)
# ──────────────────────────────────────────────
echo ""
echo -e "${BOLD}Step 5: Tool call — list_object_types${RESET}"
info "POST $SUBMIT_URL → mcp-server → objects-service"

curl -s --connect-timeout 10 "$SUBMIT_URL" \
    -H "Content-Type: application/json" \
    -d '{
        "jsonrpc": "2.0",
        "id": 3,
        "method": "tools/call",
        "params": {
            "name": "list_object_types",
            "arguments": {}
        }
    }' > /dev/null

# Wait for SSE response (may take longer due to DB query)
sleep 2

LIST_RESPONSE=$(extract_json "$SSE_OUT" '"id":3')

ERROR_CODE=$(echo "$LIST_RESPONSE" | jq -r '.error.code // empty' 2>/dev/null || echo "")
CONTENT_COUNT=$(echo "$LIST_RESPONSE" | jq '.result.content | length' 2>/dev/null || echo "0")

if [ -n "$ERROR_CODE" ]; then
    ERROR_MSG=$(echo "$LIST_RESPONSE" | jq -r '.error.message // empty' 2>/dev/null || echo "")
    fail "Tool call error (code=$ERROR_CODE): $ERROR_MSG"
elif [ "$CONTENT_COUNT" -gt 0 ] 2>/dev/null; then
    TYPE_NAMES=$(echo "$LIST_RESPONSE" | jq -r '.result.content[0].structuredContent // empty' 2>/dev/null || echo "")

    if [ "$TYPE_NAMES" != "null" ] && [ -n "$TYPE_NAMES" ]; then
        TYPE_COUNT=$(echo "$TYPE_NAMES" | jq 'length' 2>/dev/null || echo "?")
        pass "list_object_types returned $TYPE_COUNT types (objects-service chain verified)"

        # Show first few type names as sample
        SAMPLES=$(echo "$TYPE_NAMES" | jq -r '.[0:3][] | .name // .type_key // empty' 2>/dev/null || echo "")
        if [ -n "$SAMPLES" ]; then
            info "Sample types:"
            echo "$SAMPLES" | while read name; do info "- $name"; done
        fi
    else
        fail "Tool returned content but could not parse object type list"
        info "Response:"
        echo "$LIST_RESPONSE" | head -3 | sed 's/^/      /'
    fi
else
    fail "list_object_types returned empty result"
    if [ -n "$LIST_RESPONSE" ]; then
        info "Response:"
        echo "$LIST_RESPONSE" | head -3 | sed 's/^/      /'
    fi
fi

# ──────────────────────────────────────────────
# Step 6: Call get_object_type (by ID)
# ──────────────────────────────────────────────
echo ""
echo -e "${BOLD}Step 6: Tool call — get_object_type${RESET}"
info "POST $SUBMIT_URL → mcp-server → objects-service"

FIRST_TYPE_ID=$(echo "$TYPE_NAMES" | jq -r '.[0].id // empty' 2>/dev/null || echo "")

if [ -n "$FIRST_TYPE_ID" ] && [ "$FIRST_TYPE_ID" != "null" ]; then
    curl -s --connect-timeout 5 "$SUBMIT_URL" \
        -H "Content-Type: application/json" \
        -d "{\"jsonrpc\":\"2.0\",\"id\":4,\"method\":\"tools/call\",\"params\":{\"name\":\"get_object_type\",\"arguments\":{\"id\":$FIRST_TYPE_ID}}}" > /dev/null

    sleep 1

    GET_RESPONSE=$(grep "^data:" "$SSE_OUT" | grep '"id":4' | tail -1 || true)
    TYPE_NAME=$(echo "$GET_RESPONSE" | jq -r '.result.content[0].structuredContent.name // empty' 2>/dev/null || echo "")

    if [ -n "$TYPE_NAME" ] && [ "$TYPE_NAME" != "null" ]; then
        pass "get_object_type(id=$FIRST_TYPE_ID) → $TYPE_NAME"
    else
        fail "get_object_type did not return a name"
        info "Response:"
        echo "$GET_RESPONSE" | head -3 | sed 's/^/      /'
    fi
else
    warn "Skipped (no types available from list)"
fi

# ──────────────────────────────────────────────
# Step 7: Verify identity chain in database
# ──────────────────────────────────────────────
echo ""
echo -e "${BOLD}Step 7: Auth-chain configuration${RESET}"
info "Checking mcp-agent user role linkage in DB..."

ROLE_COUNT=$(docker exec service-boilerplate-postgres psql -U postgres -d service_db \
    -t -A -c "SELECT COUNT(*) FROM auth_service.user_roles ur 
              JOIN auth_service.roles r ON ur.role_id = r.id 
              WHERE r.name = 'mcp-agent-read-only'" 2>/dev/null || echo "0")

if [ "$ROLE_COUNT" = "1" ]; then
    pass "mcp-agent role linked to system user (auth chain verified)"
else
    warn "mcp-agent-role not found in auth_service.user_roles (expected: 1, got: $ROLE_COUNT)"
fi

PERM_COUNT=$(docker exec service-boilerplate-postgres psql -U postgres -d service_db \
    -t -A -c "SELECT COUNT(*) FROM auth_service.permissions WHERE resource LIKE 'objects:%'" 2>/dev/null || echo "0")
info "Objects permissions in DB: $PERM_COUNT"

# ──────────────────────────────────────────────
# Summary
# ──────────────────────────────────────────────
echo ""
echo -e "${BOLD}=== Results ===${RESET}"
echo -e "  Passed: ${GREEN}$PASS${RESET}"
if [ "$FAIL" -gt 0 ]; then
    echo -e "  Failed: ${RED}$FAIL${RESET}"
else
    echo -e "  Failed: $FAIL"
fi

if [ "$FAIL" -eq 0 ]; then
    echo ""
    echo -e "${GREEN}${BOLD}All tests passed — MCP server is fully functional through the API Gateway.${RESET}"
    exit 0
else
    echo ""
    echo -e "${RED}${BOLD}Some tests failed. See output above for details.${RESET}"
    exit 1
fi
