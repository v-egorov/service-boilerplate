package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	mcpclient "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/client"
)

func newTestServer(t *testing.T, baseURL string) (*server.MCPServer, *mcpclient.ObjectsClient) {
	t.Helper()
	mcpServer := server.NewMCPServer("test-mcp", "0.0.1")
	client := mcpclient.NewObjectsClient(baseURL, 0)
	RegisterObjectTools(mcpServer, client)
	return mcpServer, client
}

func callToolRaw(t *testing.T, srv *server.MCPServer, name string, args map[string]any) mcp.JSONRPCMessage {
	t.Helper()
	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": args,
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	resp := srv.HandleMessage(context.Background(), raw)
	return resp
}

// extractStructuredContentMap extracts a map[string]any from structuredContent,
// handling both raw maps and typed structs (ListObjectsResult, GetObjectResult).
func extractStructuredContentMap(t *testing.T, sc any) map[string]any {
	t.Helper()
	if m, ok := sc.(map[string]any); ok {
		return m
	}
	// Handle typed struct results by serializing/deserializing through JSON
	data, err := json.Marshal(sc)
	if err != nil {
		t.Fatalf("failed to marshal structuredContent: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal structuredContent: %v", err)
	}
	return result
}

func getCallToolResult(t *testing.T, msg mcp.JSONRPCMessage) (*mcp.CallToolResult, bool) {
	t.Helper()
	jr, ok := msg.(mcp.JSONRPCResponse)
	if !ok {
		return nil, false
	}
	result, ok := jr.Result.(*mcp.CallToolResult)
	return result, ok
}

func TestListObjectsParamsLimitOffset(t *testing.T) {
	var capturedURL string
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"pagination":{"total":0,"limit":3,"offset":0}}`))
	}))
	defer testServer.Close()

	srv, _ := newTestServer(t, testServer.URL)

	msg := callToolRaw(t, srv, "list_objects", map[string]any{
		"object_type_id": 42,
		"limit":          3,
		"offset":         0,
	})

	result, ok := getCallToolResult(t, msg)
	if !ok {
		t.Fatal("expected JSONRPCResponse with CallToolResult")
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
	if capturedURL == "" {
		t.Fatal("expected request to be made")
	}
	if !strings.Contains(capturedURL, "limit=3") {
		t.Errorf("URL should contain 'limit=3', got: %s", capturedURL)
	}
	if !strings.Contains(capturedURL, "offset=0") {
		t.Errorf("URL should contain 'offset=0', got: %s", capturedURL)
	}
	if strings.Contains(capturedURL, "page=") || strings.Contains(capturedURL, "page_size=") {
		t.Errorf("URL should NOT contain 'page=' or 'page_size=', got: %s", capturedURL)
	}

	sc := result.StructuredContent
	if sc == nil {
		t.Fatal("expected StructuredContent to be non-nil")
	}
	scMap := extractStructuredContentMap(t, sc)
	objs, ok := scMap["objects"].([]any)
	if !ok || len(objs) != 0 {
		t.Errorf("expected objects=[] in structuredContent")
	}
}

func TestListObjectsZeroLimitDefaultsTo50(t *testing.T) {
	var capturedURL string
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"pagination":{"total":0,"limit":50,"offset":0}}`))
	}))
	defer testServer.Close()

	srv, _ := newTestServer(t, testServer.URL)

	msg := callToolRaw(t, srv, "list_objects", map[string]any{
		"object_type_id": 42,
		"limit":          0, // zero should default to 50
	})

	result, ok := getCallToolResult(t, msg)
	if !ok || result == nil {
		t.Fatal("expected valid result")
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
	if !strings.Contains(capturedURL, "limit=50") {
		t.Errorf("zero limit should default to 50, got URL: %s", capturedURL)
	}
}

func TestListObjectsNegativeLimit(t *testing.T) {
	var capturedURL string
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"pagination":{"total":0,"limit":50,"offset":0}}`))
	}))
	defer testServer.Close()

	srv, _ := newTestServer(t, testServer.URL)

	msg := callToolRaw(t, srv, "list_objects", map[string]any{
		"object_type_id": 42,
		"limit":          -5, // negative should default to 50
	})

	result, ok := getCallToolResult(t, msg)
	if !ok || result == nil {
		t.Fatal("expected valid result")
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
	if !strings.Contains(capturedURL, "limit=50") {
		t.Errorf("negative limit should default to 50, got URL: %s", capturedURL)
	}
}

func TestListObjectsNegativeOffsetIgnored(t *testing.T) {
	var capturedURL string
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"pagination":{"total":0,"limit":50,"offset":0}}`))
	}))
	defer testServer.Close()

	srv, _ := newTestServer(t, testServer.URL)

	msg := callToolRaw(t, srv, "list_objects", map[string]any{
		"object_type_id": 42,
		"offset":         -3, // negative offset should default to 0
	})

	result, ok := getCallToolResult(t, msg)
	if !ok || result == nil {
		t.Fatal("expected valid result")
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
	if !strings.Contains(capturedURL, "offset=0") {
		t.Errorf("negative offset should default to 0, got URL: %s", capturedURL)
	}
}

func TestListObjectsNilPaginationMeta(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Response without pagination field (edge case: old objects-service version)
		w.Write([]byte(`{"data":[{"id":1,"name":"Test"}]}`))
	}))
	defer testServer.Close()

	srv, _ := newTestServer(t, testServer.URL)

	msg := callToolRaw(t, srv, "list_objects", map[string]any{
		"object_type_id": 42,
	})

	result, ok := getCallToolResult(t, msg)
	if !ok || result == nil {
		t.Fatal("expected valid result")
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}

	sc := result.StructuredContent
	if sc == nil {
		t.Fatal("expected StructuredContent to be non-nil")
	}
	scMap := extractStructuredContentMap(t, sc)

	// When pagination metadata is unavailable (nil), "pagination" key should not appear or be nil
	paginationRaw, hasPagination := scMap["pagination"]
	if hasPagination && paginationRaw != nil {
		t.Errorf("when pagination metadata unavailable, 'pagination' key should be omitted or nil")
	}

	objs, ok := scMap["objects"].([]any)
	if !ok || len(objs) != 1 {
		t.Errorf("expected 1 object in structuredContent")
	}
}

func TestListObjectsPaginationMetadataInStructuredContent(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":1,"name":"Widget"},{"id":2,"name":"Gadget"}],"pagination":{"total":7,"limit":3,"offset":0}}`))
	}))
	defer testServer.Close()

	srv, _ := newTestServer(t, testServer.URL)

	msg := callToolRaw(t, srv, "list_objects", map[string]any{
		"object_type_id": 42,
		"limit":          3,
	})

	result, ok := getCallToolResult(t, msg)
	if !ok || result == nil {
		t.Fatal("expected valid result")
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}

	sc := result.StructuredContent
	if sc == nil {
		t.Fatal("expected StructuredContent to be non-nil")
	}
	scMap := extractStructuredContentMap(t, sc)

	objs, ok := scMap["objects"].([]any)
	if !ok || len(objs) != 2 {
		t.Errorf("expected 2 objects in structuredContent")
	}

	paginationRaw, hasPagination := scMap["pagination"]
	if !hasPagination {
		t.Fatal("expected 'pagination' key in structuredContent")
	}
	paginationMap, ok := paginationRaw.(map[string]any)
	if !ok {
		t.Fatalf("expected pagination to be map[string]any, got %T", paginationRaw)
	}

	totalVal, ok := paginationMap["total"].(float64)
	if !ok || int(totalVal) != 7 {
		t.Errorf("expected pagination.total=7, got: %+v", paginationMap["total"])
	}
	limitVal, ok := paginationMap["limit"].(float64)
	if !ok || int(limitVal) != 3 {
		t.Errorf("expected pagination.limit=3, got: %+v", paginationMap["limit"])
	}
	offsetVal, ok := paginationMap["offset"].(float64)
	if !ok || int(offsetVal) != 0 {
		t.Errorf("expected pagination.offset=0, got: %+v", paginationMap["offset"])
	}
}

// Note: list_object_types is registered in type_tools.go (requires logrus.Logger),
// so it's tested via e2e/test-mcp-e2e.sh instead of unit tests here.

func TestGetObjectReturnsWrappedObject(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"id":1,"name":"Widget","public_id":"abc-123"}}`))
	}))
	defer testServer.Close()

	srv, _ := newTestServer(t, testServer.URL)

	msg := callToolRaw(t, srv, "get_object", map[string]any{
		"public_id": "abc-123",
	})
	result, ok := getCallToolResult(t, msg)
	if !ok || result == nil {
		t.Fatal("expected valid result")
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}

	sc := result.StructuredContent
	if sc == nil {
		t.Fatal("expected StructuredContent")
	}
	scMap := extractStructuredContentMap(t, sc)

	objData, ok := scMap["object"].(map[string]any)
	if !ok || objData["name"] != "Widget" {
		t.Errorf("expected object.name='Widget'")
	}
}

func TestListObjectsMissingObjectTypeID(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[]}`))
	}))
	defer testServer.Close()

	srv, _ := newTestServer(t, testServer.URL)

	msg := callToolRaw(t, srv, "list_objects", map[string]any{
		// object_type_id missing → should return error
	})
	result, ok := getCallToolResult(t, msg)
	if !ok || result == nil {
		t.Fatal("expected valid result")
	}
	if !result.IsError {
		t.Fatal("expected error when object_type_id is missing")
	}
}

func TestListObjectsWithDefaultLimit(t *testing.T) {
	var capturedURL string
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"pagination":{"total":0,"limit":50,"offset":0}}`))
	}))
	defer testServer.Close()

	srv, _ := newTestServer(t, testServer.URL)

	msg := callToolRaw(t, srv, "list_objects", map[string]any{
		"object_type_id": 42,
		// limit not provided → should default to 50
	})

	result, ok := getCallToolResult(t, msg)
	if !ok || result == nil {
		t.Fatal("expected valid result")
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
	if !strings.Contains(capturedURL, "limit=50") {
		t.Errorf("missing limit should default to 50, got URL: %s", capturedURL)
	}
}

func TestListObjectsCustomLimitAndOffset(t *testing.T) {
	var capturedURL string
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"pagination":{"total":100,"limit":15,"offset":30}}`))
	}))
	defer testServer.Close()

	srv, _ := newTestServer(t, testServer.URL)

	msg := callToolRaw(t, srv, "list_objects", map[string]any{
		"object_type_id": 42,
		"limit":          15,
		"offset":         30,
	})

	result, ok := getCallToolResult(t, msg)
	if !ok || result == nil {
		t.Fatal("expected valid result")
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
	if !strings.Contains(capturedURL, "limit=15") || !strings.Contains(capturedURL, "offset=30") {
		t.Errorf("expected limit=15 and offset=30 in URL, got: %s", capturedURL)
	}

	scMap := extractStructuredContentMap(t, result.StructuredContent)
	paginationRaw := scMap["pagination"]
	if paginationRaw == nil {
		t.Fatal("expected 'pagination' key in structuredContent")
	}
	paginationMap, ok := paginationRaw.(map[string]any)
	if !ok {
		t.Fatalf("expected pagination to be map[string]any, got %T", paginationRaw)
	}
	if int(paginationMap["total"].(float64)) != 100 {
		t.Errorf("expected total=100 in pagination metadata")
	}
}

func TestGetObjectMissingPublicID(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{}}`))
	}))
	defer testServer.Close()

	srv, _ := newTestServer(t, testServer.URL)

	msg := callToolRaw(t, srv, "get_object", map[string]any{}) // public_id missing
	result, ok := getCallToolResult(t, msg)
	if !ok || result == nil {
		t.Fatal("expected valid result")
	}
	if !result.IsError {
		t.Fatal("expected error when public_id is missing")
	}
}

func TestGetObjectStructuredContentIsObjectNotArray(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"id":1,"name":"Widget"}}`))
	}))
	defer testServer.Close()

	srv, _ := newTestServer(t, testServer.URL)

	msg := callToolRaw(t, srv, "get_object", map[string]any{
		"public_id": "abc-123",
	})
	result, ok := getCallToolResult(t, msg)
	if !ok || result == nil {
		t.Fatal("expected valid result")
	}

	sc := result.StructuredContent
	// Marshal/deserialize to check the actual JSON shape (typed structs become maps after round-trip)
	data, err := json.Marshal(sc)
	if err != nil {
		t.Fatalf("failed to marshal structuredContent: %v", err)
	}
	var decoded any
	json.Unmarshal(data, &decoded)
	if _, isArray := decoded.([]any); isArray {
		t.Error("get_object structuredContent should be a JSON object, not an array")
	}
}

// TestPaginationMetaJSONSerialization verifies PaginationMeta serializes correctly.
func TestPaginationMetaJSONSerialization(t *testing.T) {
	p := &PaginationMeta{Total: 42, Limit: 10, Offset: 5}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("failed to marshal PaginationMeta: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if int(decoded["total"].(float64)) != 42 {
		t.Errorf("expected total=42")
	}
	if int(decoded["limit"].(float64)) != 10 {
		t.Errorf("expected limit=10")
	}
	if int(decoded["offset"].(float64)) != 5 {
		t.Errorf("expected offset=5")
	}
}

// TestListObjectsResultJSONSerialization verifies ListObjectsResult serializes correctly.
func TestListObjectsResultJSONSerialization(t *testing.T) {
	result := ListObjectsResult{
		Items:      []map[string]interface{}{{"id": 1, "name": "Widget"}},
		Pagination: &PaginationMeta{Total: 7, Limit: 3, Offset: 0},
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal ListObjectsResult: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	objs, ok := decoded["objects"].([]any)
	if !ok || len(objs) != 1 {
		t.Errorf("expected objects array with 1 item")
	}

	pagination, ok := decoded["pagination"].(map[string]any)
	if !ok {
		t.Fatal("expected pagination key in serialized result")
	}
	if int(pagination["total"].(float64)) != 7 {
		t.Errorf("expected total=7")
	}
}

// TestListObjectsResultNilPaginationOmitempty verifies omitempty works.
func TestListObjectsResultNilPaginationOmitempty(t *testing.T) {
	result := ListObjectsResult{
		Items:      []map[string]interface{}{{"id": 1, "name": "Widget"}},
		Pagination: nil,
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, hasPagination := decoded["pagination"]; hasPagination {
		t.Error("expected 'pagination' key to be omitted when nil (omitempty)")
	}
}
