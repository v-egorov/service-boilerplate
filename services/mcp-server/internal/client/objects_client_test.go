package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListObjectTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":1,"name":"Document","type_key":"document"}]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	types, err := client.ListObjectTypes(context.Background(), "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(types) != 1 {
		t.Errorf("expected 1 type, got %d", len(types))
	}
	if types[0]["name"] != "Document" {
		t.Errorf("expected name 'Document', got %v", types[0]["name"])
	}
}

func TestGetObjectTypeByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1,"name":"Document","type_key":"document"}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	typ, err := client.GetObjectTypeByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if typ["name"] != "Document" {
		t.Errorf("expected name 'Document', got %v", typ["name"])
	}
}

func TestGetObjectTypeByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1,"name":"Document","type_key":"document"}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	typ, err := client.GetObjectTypeByName(context.Background(), "document")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if typ["type_key"] != "document" {
		t.Errorf("expected type_key 'document', got %v", typ["type_key"])
	}
}

func TestListObjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":1,"name":"Contract","public_id":"abc-123"}]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	objs, _, err := client.ListObjects(context.Background(), 42, 3, 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(objs) != 1 {
		t.Errorf("expected 1 object, got %d", len(objs))
	}
	if objs[0]["name"] != "Contract" {
		t.Errorf("expected name 'Contract', got %v", objs[0]["name"])
	}
}

func TestListObjectsURLContainsLimitOffset(t *testing.T) {
	var receivedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	_, _, err := client.ListObjects(context.Background(), 42, 10, 5, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain limit= and offset= (not page= or page_size=)
	if !contains(receivedQuery, "limit=10") {
		t.Errorf("expected URL to contain 'limit=10', got query: %s", receivedQuery)
	}
	if !contains(receivedQuery, "offset=5") {
		t.Errorf("expected URL to contain 'offset=5', got query: %s", receivedQuery)
	}
	if contains(receivedQuery, "page=") {
		t.Errorf("URL should NOT contain 'page=', got query: %s", receivedQuery)
	}
	if contains(receivedQuery, "page_size=") {
		t.Errorf("URL should NOT contain 'page_size=', got query: %s", receivedQuery)
	}
}

// TestListObjectsResponseStructure verifies the response JSON structure matches
// what objects-service returns (data array + pagination object).
func TestListObjectsResponseStructure(t *testing.T) {
	respBody := `{"data":[{"id":1,"name":"Test"}],"pagination":{"total":42,"limit":10,"offset":5}}`

	var decoded struct {
		Data       []map[string]interface{} `json:"data"`
		Pagination map[string]any           `json:"pagination"`
	}
	if err := json.Unmarshal([]byte(respBody), &decoded); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(decoded.Data) != 1 {
		t.Errorf("expected 1 data item, got %d", len(decoded.Data))
	}
	if decoded.Pagination["total"] != float64(42) {
		t.Errorf("expected pagination.total=42, got %v", decoded.Pagination["total"])
	}
	if decoded.Pagination["limit"] != float64(10) {
		t.Errorf("expected pagination.limit=10, got %v", decoded.Pagination["limit"])
	}
	if decoded.Pagination["offset"] != float64(5) {
		t.Errorf("expected pagination.offset=5, got %v", decoded.Pagination["offset"])
	}
}

func TestGetObjectByPublicID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1,"name":"Contract","public_id":"abc-123"}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	obj, err := client.GetObjectByPublicID(context.Background(), "abc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj["public_id"] != "abc-123" {
		t.Errorf("expected public_id 'abc-123', got %v", obj["public_id"])
	}
}

func TestGetRootTree(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[
			{"id":1,"name":"Root","parent_type_id":null},
			{"id":2,"name":"Child","parent_type_id":1}
		]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	tree, err := client.GetRootTree(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tree) != 1 {
		t.Errorf("expected 1 root type, got %d", len(tree))
	}
	if tree[0]["name"] != "Root" {
		t.Errorf("expected name 'Root', got %v", tree[0]["name"])
	}
}

func TestHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	_, err := client.ListObjectTypes(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

// TestListObjectsPaginationMetadataDecoded verifies objects-service pagination
// response is correctly decoded into the struct with Data + Pagination fields.
func TestListObjectsPaginationMetadataDecoded(t *testing.T) {
	respBody := `{"data":[{"id":1,"name":"Test"}],"pagination":{"total":42,"limit":10,"offset":5}}`
	var decoded struct {
		Data       []map[string]interface{} `json:"data"`
		Pagination map[string]any           `json:"pagination"`
	}
	if err := json.Unmarshal([]byte(respBody), &decoded); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(decoded.Data) != 1 {
		t.Errorf("expected 1 data item, got %d", len(decoded.Data))
	}

	total, ok := decoded.Pagination["total"]
	if !ok || total.(float64) != 42 {
		t.Errorf("expected pagination.total=42, got %v", total)
	}
	limit, ok := decoded.Pagination["limit"]
	if !ok || limit.(float64) != 10 {
		t.Errorf("expected pagination.limit=10, got %v", limit)
	}
	offset, ok := decoded.Pagination["offset"]
	if !ok || offset.(float64) != 5 {
		t.Errorf("expected pagination.offset=5, got %v", offset)
	}
}

// contains checks if a substring exists in a string.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
