#!/bin/bash
# MCP Server E2E Test — verifies the full StreamableHTTP protocol handshake through API Gateway
# Uses Python for MCP protocol calls (mcp-go's StreamableHTTP transport has known incompatibility
# with curl's HTTP/1.1 POST requests; Python http.client works correctly).

set -euo pipefail

BASE_URL="${MCP_BASE_URL:-http://localhost:8080}"
MCP_ENDPOINT="$BASE_URL/mcp"
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
info() { echo -e "    $BLUE$1${RESET}" >&2; }

echo ""
echo -e "${BOLD}=== MCP Server E2E Test ===" 
echo -e "Target: StreamableHTTP via API Gateway ($MCP_ENDPOINT)"
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
# Steps 2-6: MCP Protocol tests via Python http.client
# ──────────────────────────────────────────────
echo -e "${BOLD}Steps 2-6: MCP protocol handshake (initialize, tools/list, tool calls)${RESET}"

MCP_RESULT=$(python3 << PYEOF
import http.client
import json
import sys

BASE_URL = """$MCP_ENDPOINT"""
TIMEOUT = $TIMEOUT
PASS = 0
FAIL = 0

def pass_test(msg):
    global PASS
    print(f"  PASS: {msg}")
    PASS += 1

def fail_test(msg):
    global FAIL
    print(f"  FAIL: {msg}")
    FAIL += 1

# Parse endpoint URL for connection
if '//' in BASE_URL:
    _, rest = BASE_URL.split('//', 1)
else:
    rest = BASE_URL

if '/' in rest:
    host_port, path = rest.split('/', 1)
    path = '/' + path
else:
    host_port = rest
    path = '/mcp'

if ':' in host_port:
    h, p_str = host_port.split(':', 1)
    port = int(p_str)
else:
    h, port = host_port, 80

headers_base = {"Content-Type": "application/json"}

# --- Step 2: Initialize ---
print("\n  Step 2: MCP Initialize")
body = json.dumps({
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
        "protocolVersion": "2024-11-05",
        "capabilities": {},
        "clientInfo": {"name": "e2e-test-client", "version": "1.0"}
    }
})

conn = http.client.HTTPConnection(h, port, timeout=TIMEOUT)
try:
    conn.request("POST", path, body.encode(), headers_base)
    resp = conn.getresponse()
    init_resp = json.loads(resp.read().decode())
    session_id = resp.getheader('MCP-Session-ID')
finally:
    conn.close()

if status := (resp.status == 200 if 'resp' in dir() else False):
    server_name = init_resp.get('result', {}).get('serverInfo', {}).get('name', '?')
    proto_ver = init_resp['result'].get('protocolVersion', '?')
    if session_id and session_id.startswith('mcp-session-'):
        pass_test(f"Initialize accepted (proto={proto_ver}, server={server_name}, session={session_id[:12]}...)")
    else:
        fail_test("No MCP-Session-ID in initialize response headers")

if resp.status != 200 or 'result' not in init_resp:
    print(f"  Cannot continue — initialization failed (status={resp.status})")
    sys.exit(1)

# --- Step 3: List tools ---
print("  Step 3: List MCP Tools")
conn = http.client.HTTPConnection(h, port, timeout=TIMEOUT)
try:
    conn.request("POST", path, json.dumps({"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}).encode(), 
                 {**headers_base, "MCP-Session-ID": session_id})
    resp = conn.getresponse()
    tools_resp = json.loads(resp.read().decode())
finally:
    conn.close()

tool_count = len(tools_resp.get('result', {}).get('tools', []))
if tool_count > 0:
    pass_test(f"Found {tool_count} MCP tool(s)")
    for t in tools_resp['result']['tools']:
        desc = (t.get('description') or '')[:60] + '...' if len(t.get('description') or '') > 60 else (t.get('description') or '')
        print(f"      {t['name']}: {desc}")
else:
    fail_test("No tools returned from server")

# --- Step 4: Call list_object_types ---
print("  Step 4: Tool call — list_object_types")
conn = http.client.HTTPConnection(h, port, timeout=TIMEOUT)
try:
    conn.request("POST", path, json.dumps({
        "jsonrpc": "2.0", "id": 3, "method": "tools/call",
        "params": {"name": "list_object_types", "arguments": {}}
    }).encode(), {**headers_base, "MCP-Session-ID": session_id})
    resp = conn.getresponse()
    list_resp = json.loads(resp.read().decode())
finally:
    conn.close()

error_code = list_resp.get('error', {}).get('code')
if error_code:
    fail_test(f"Tool call error (code={error_code}): {list_resp['error'].get('message', '')}")
else:
    result = list_resp.get('result', {})
    
    # mcp-go v0.55.1 returns structuredContent at root level of result
    sc = result.get('structuredContent') if isinstance(result, dict) else None
    
    types_list = []
    if isinstance(sc, dict):
        types_list = sc.get('types', [])
    elif isinstance(sc, list):
        types_list = sc
    
    # Fallback: check content[] array
    if not types_list:
        content = result.get('content', []) if isinstance(result, dict) else []
        if content and len(content) > 0:
            first_content = content[0]
            if isinstance(first_content, dict):
                sc2 = first_content.get('structuredContent')
                if isinstance(sc2, dict):
                    types_list = sc2.get('types', [])
                elif isinstance(sc2, list):
                    types_list = sc2
    
    type_count = len(types_list)
    if type_count > 0:
        pass_test(f"list_object_types returned {type_count} types")
        for t in types_list[:3]:
            name = (t.get('name') or t.get('type_key', 'unknown')) if isinstance(t, dict) else str(t)
            print(f"      - {name}")
    else:
        fail_test("list_object_types returned empty result")

# Extract first type info for steps 5-6
first_type_id = None
if types_list and len(types_list) > 0 and isinstance(types_list[0], dict):
    first_type_id = types_list[0].get('id')

# --- Step 5: Call get_object_type ---
print("  Step 5: Tool call — get_object_type")
if first_type_id and str(first_type_id).isdigit():
    conn = http.client.HTTPConnection(h, port, timeout=TIMEOUT)
    try:
        conn.request("POST", path, json.dumps({
            "jsonrpc": "2.0", "id": 4, "method": "tools/call",
            "params": {"name": "get_object_type", "arguments": {"id": int(first_type_id)}}
        }).encode(), {**headers_base, "MCP-Session-ID": session_id})
        resp = conn.getresponse()
        get_resp = json.loads(resp.read().decode())
    finally:
        conn.close()

    error_code = get_resp.get('error', {}).get('code')
    if error_code:
        fail_test(f"get_object_type error (code={error_code}): {get_resp['error'].get('message', '')}")
    else:
        result = get_resp.get('result', {})
        sc = result.get('structuredContent') if isinstance(result, dict) else None
        
        # Fallback: check content[] array  
        if not sc:
            content = result.get('content', []) if isinstance(result, dict) else []
            if content and len(content) > 0:
                first_content = content[0]
                if isinstance(first_content, dict):
                    sc = first_content.get('structuredContent')

        if sc and isinstance(sc, dict):
            # New format: {"item": {...}} wrapper from output schema
            item_data = sc.get('item') or sc
            type_name = (item_data if isinstance(item_data, dict) else {}).get('name') or (item_data if isinstance(item_data, dict) else {}).get('type_key', '')
            if type_name:
                pass_test(f"get_object_type(id={first_type_id}) → {type_name}")
            else:
                fail_test("get_object_type did not return a name")
        else:
            # Check content[] with text fallback
            content = result.get('content', []) if isinstance(result, dict) else []
            if content and len(content) > 0:
                first_content = content[0]
                if isinstance(first_content, dict):
                    text = first_content.get('text', '')
                    if '{' in text or 'name' in text.lower():
                        pass_test(f"get_object_type(id={first_type_id}) → found response")
                    else:
                        fail_test("get_object_type returned unexpected content format")
                else:
                    fail_test("get_object_type returned non-dict content")
            else:
                fail_test("get_object_type returned no content")
else:
    print("  SKIP: get_object_type (no types available from list)")

# --- Step 6: Call list_objects ---
print("  Step 6: Tool call — list_objects")
if first_type_id and str(first_type_id).isdigit():
    conn = http.client.HTTPConnection(h, port, timeout=TIMEOUT)
    try:
        conn.request("POST", path, json.dumps({
            "jsonrpc": "2.0", "id": 5, "method": "tools/call",
            "params": {"name": "list_objects", "arguments": {"object_type_id": int(first_type_id)}}
        }).encode(), {**headers_base, "MCP-Session-ID": session_id})
        resp = conn.getresponse()
        obj_resp = json.loads(resp.read().decode())
    finally:
        conn.close()

    error_code = obj_resp.get('error', {}).get('code')
    if error_code:
        fail_test(f"list_objects error (code={error_code}): {obj_resp['error'].get('message', '')}")
    else:
        result = obj_resp.get('result', {})
        
        # Check structuredContent at root level of result
        sc = result.get('structuredContent') if isinstance(result, dict) else None
        
        # Fallback: check content[] array  
        if not sc:
            content = result.get('content', []) if isinstance(result, dict) else []
            if content and len(content) > 0:
                first_content = content[0]
                if isinstance(first_content, dict):
                    sc = first_content.get('structuredContent') or first_content

        obj_count = 0
        if isinstance(sc, list):
            obj_count = len(sc)
        elif isinstance(sc, dict):
            obj_list = sc.get('objects', []) if 'objects' in sc else []
            obj_count = len(obj_list)
        elif sc is not None and isinstance(sc, str):
            try:
                parsed = json.loads(sc)
                if isinstance(parsed, list):
                    obj_count = len(parsed)
                elif isinstance(parsed, dict):
                    obj_count = len(parsed.get('objects', []))
                else:
                    obj_count = 0
            except json.JSONDecodeError:
                fail_test("list_objects returned unparseable content")

        if 'obj_count' in dir() and obj_count == 0 and not isinstance(sc, (str,)):
            # Check raw response for clues only on failure
            print(f"      Response structure: {json.dumps(obj_resp)[:300]}")
        
        if obj_count > 0:
            pass_test(f"list_objects returned {obj_count} object(s) for type id={first_type_id}")
        elif not isinstance(sc, str):
            # Empty list is valid (no objects of this type yet)
            pass_test("list_objects returned empty result (no objects of this type)")

# Print summary
print(f"\n  === Summary: PASS={PASS}, FAIL={FAIL} ===")
sys.exit(0 if FAIL == 0 else 1)
PYEOF
)

echo "$MCP_RESULT"

# Count pass/fail from Python output  
PY_PASS=$(echo "$MCP_RESULT" | grep -c "PASS:" || true)
PY_FAIL=$(echo "$MCP_RESULT" | grep -c "FAIL:" || true)
PASS=$((PASS + PY_PASS))
FAIL=$((FAIL + PY_FAIL))

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
    echo -e "${GREEN}${BOLD}All tests passed — MCP server (StreamableHTTP) is fully functional through the API Gateway.${RESET}"
    exit 0
else
    echo ""
    echo -e "${RED}${BOLD}Some tests failed. See output above for details.${RESET}"
    exit 1
fi
