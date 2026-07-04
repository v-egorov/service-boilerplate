package client

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	types, err := client.ListObjectTypes(context.Background(), "")
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
	objs, err := client.ListObjects(context.Background(), 42, 1, 10)
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
	_, err := client.ListObjectTypes(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}
