package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListObjectsForwardsIdentityHeaders(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "test-user-123")
	hdr.Set("X-User-Email", "test@example.com")
	hdr.Set("X-User-Roles", "admin,user")

	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.ListObjects(ctx, 42, 1, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "test-user-123")
	assertHeader(t, received, "X-User-Email", "test@example.com")
	assertHeader(t, received, "X-User-Roles", "admin,user")
}

func TestListObjectsDoesNotForwardNonIdentityHeaders(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "test-user-123")
	hdr.Set("Authorization", "Bearer secret-token")
	hdr.Set("Cookie", "session=abc123")
	hdr.Set("X-Custom-Header", "secret-value")

	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.ListObjects(ctx, 42, 1, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "test-user-123")
	// Non-identity headers must not be forwarded (Go's User-Agent is added by transport, not our code)
	assertNilOrEmpty(t, received, "Authorization")
	assertNilOrEmpty(t, received, "Cookie")
	assertNilOrEmpty(t, received, "X-Custom-Header")
}

func TestListObjectsProceedsWithoutIdentityHeaders(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	// No WithIdentity call — context has no identity
	ctx := context.Background()
	_, err := client.ListObjects(ctx, 42, 1, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
	assertNilOrEmpty(t, received, "X-User-Email")
	assertNilOrEmpty(t, received, "X-User-Roles")
}

func TestWithIdentityNilSafe(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := WithIdentity(context.Background(), nil)
	_, err := client.ListObjects(ctx, 42, 1, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
}

func assertHeader(t *testing.T, hdr http.Header, key, expected string) {
	t.Helper()
	got := hdr.Get(key)
	if got != expected {
		t.Errorf("header %q: expected %q, got %q", key, expected, got)
	}
}

func assertNilOrEmpty(t *testing.T, hdr http.Header, key string) {
	t.Helper()
	got := hdr.Get(key)
	if got != "" {
		t.Errorf("header %q: expected empty, got %q", key, got)
	}
}

// TestIdentityFromContextNil verifies IdentityFromContext returns nil for a plain context.
func TestIdentityFromContextNil(t *testing.T) {
	got := IdentityFromContext(context.Background())
	if got != nil {
		t.Errorf("expected nil identity from plain context, got %v", got)
	}
}

// TestGetRootTreeForwardsIdentity ensures the resource path also forwards headers.
func TestGetRootTreeForwardsIdentity(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":1,"name":"Root","parent_type_id":null}]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "tree-user")
	hdr.Set("Authorization", "Bearer secret")
	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.GetRootTree(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "tree-user")
	// Non-identity headers must not leak (Go's User-Agent is added by transport)
	assertNilOrEmpty(t, received, "Authorization")
}

// TestHealthCheckURLUnchanged verifies HealthCheckURL is unaffected by identity changes.
func TestHealthCheckURLUnchanged(t *testing.T) {
	client := NewObjectsClient("http://objects:8080", 0)
	url := client.HealthCheckURL()
	if url != "http://objects:8080" {
		t.Errorf("expected 'http://objects:8080', got %q", url)
	}
}

// TestForwardHeadersList verifies the forward list is exactly the expected three headers.
func TestForwardHeadersList(t *testing.T) {
	expected := []string{"X-User-ID", "X-User-Email", "X-User-Roles"}
	if len(forwardHeaders) != 3 {
		t.Fatalf("expected 3 forward headers, got %d", len(forwardHeaders))
	}
	for i, h := range expected {
		if forwardHeaders[i] != h {
			t.Errorf("forwardHeaders[%d]: expected %q, got %q", i, h, forwardHeaders[i])
		}
	}
}

// TestGetObjectTypeByIDForwardsIdentity ensures all data methods forward headers.
func TestGetObjectTypeByIDForwardsIdentity(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1,"name":"Doc"}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "type-user")
	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.GetObjectTypeByID(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "type-user")
}

// TestGetObjectTypeByNameForwardsIdentity ensures name-based lookup also forwards headers.
func TestGetObjectTypeByNameForwardsIdentity(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1,"name":"Doc"}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "name-user")
	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.GetObjectTypeByName(ctx, "document")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "name-user")
}

// TestGetObjectTypeTreeForwardsIdentity ensures tree lookup also forwards headers.
func TestGetObjectTypeTreeForwardsIdentity(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "tree-id-user")
	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.GetObjectTypeTree(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "tree-id-user")
}

// TestGetObjectByPublicIDForwardsIdentity ensures object lookup also forwards headers.
func TestGetObjectByPublicIDForwardsIdentity(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1,"name":"Obj"}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "obj-user")
	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.GetObjectByPublicID(ctx, "pub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "obj-user")
}

// TestIdentityHeadersNotWholesaleCopied verifies that arbitrary headers in the
// inbound request are never forwarded wholesale to objects-service.
func TestIdentityHeadersNotWholesaleCopied(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "user-a")
	hdr.Set("Authorization", "Bearer secret-token")
	hdr.Set("Cookie", "session=abc123")
	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.ListObjects(ctx, 42, 1, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "user-a")
	assertNilOrEmpty(t, received, "Authorization")
	assertNilOrEmpty(t, received, "Cookie")
}

// TestListObjectsMultipleIdentityHeaders verifies all three identity headers forward.
func TestListObjectsMultipleIdentityHeaders(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "multi-user")
	hdr.Set("X-User-Email", "multi@example.com")
	hdr.Set("X-User-Roles", "admin,user,viewer")
	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.ListObjects(ctx, 42, 1, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "multi-user")
	assertHeader(t, received, "X-User-Email", "multi@example.com")
	assertHeader(t, received, "X-User-Roles", "admin,user,viewer")
}

// TestGetRootTreeWithNilContext verifies GetRootTree works without identity.
func TestGetRootTreeWithNilContext(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":1,"name":"Root","parent_type_id":null}]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := context.Background()
	_, err := client.GetRootTree(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
	assertNilOrEmpty(t, received, "X-User-Email")
	assertNilOrEmpty(t, received, "X-User-Roles")
}

// TestGetRootTreeWithNilIdentity verifies WithIdentity(nil) is nil-safe.
func TestGetRootTreeWithNilIdentity(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":1,"name":"Root","parent_type_id":null}]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := WithIdentity(context.Background(), nil)
	_, err := client.GetRootTree(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
}

// TestGetObjectTypeByIDWithNilContext verifies GetObjectByPublicID works without identity.
func TestGetObjectByPublicIDWithNilContext(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := context.Background()
	_, err := client.GetObjectByPublicID(ctx, "pub-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
}

// TestGetObjectTypeByNameWithNilContext verifies GetObjectByPublicID works without identity.
func TestGetObjectTypeByNameWithNilContext(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := context.Background()
	_, err := client.GetObjectTypeByName(ctx, "document")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
}

// TestListObjectsWithNilContext verifies ListObjects works without identity.
func TestListObjectsWithNilContext(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := context.Background()
	_, err := client.ListObjects(ctx, 42, 1, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
}

// TestGetObjectTypeByIDWithNilContext verifies GetObjectTypeByID works without identity.
func TestGetObjectTypeByIDWithNilContext(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := context.Background()
	_, err := client.GetObjectTypeByID(ctx, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
}

// TestGetObjectTypeTreeWithNilContext verifies GetObjectTypeTree works without identity.
func TestGetObjectTypeTreeWithNilContext(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := context.Background()
	_, err := client.GetObjectTypeTree(ctx, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
}

// TestListObjectTypesForwardsIdentity ensures ListObjectTypes forwards headers.
func TestListObjectTypesForwardsIdentity(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "list-types-user")
	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.ListObjectTypes(ctx, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "list-types-user")
}

// TestListObjectsWithNilContext verifies ListObjects works without identity.
func TestListObjectsWithNilIdentityValue(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := WithIdentity(context.Background(), nil)
	_, err := client.ListObjects(ctx, 42, 1, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
}

// TestListObjectTypesWithNilContext ensures ListObjectTypes works without identity.
func TestListObjectTypesWithNilContext(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := context.Background()
	_, err := client.ListObjectTypes(ctx, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
}

// TestGetObjectTypeByIDWithNilContext ensures GetObjectTypeByID works without identity.
func TestGetObjectTypeByIDWithNilIdentityValue(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := WithIdentity(context.Background(), nil)
	_, err := client.GetObjectTypeByID(ctx, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
}

// TestGetObjectTypeByNameWithNilContext ensures GetObjectTypeByName works without identity.
func TestGetObjectTypeByNameWithNilIdentityValue(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := WithIdentity(context.Background(), nil)
	_, err := client.GetObjectTypeByName(ctx, "document")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
}

// TestGetObjectTypeTreeWithNilIdentity ensures GetObjectTypeTree works without identity.
func TestGetObjectTypeTreeWithNilIdentityValue(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := WithIdentity(context.Background(), nil)
	_, err := client.GetObjectTypeTree(ctx, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
}

// TestGetObjectByPublicIDWithNilIdentity ensures GetObjectByPublicID works without identity.
func TestGetObjectByPublicIDWithNilIdentityValue(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := WithIdentity(context.Background(), nil)
	_, err := client.GetObjectByPublicID(ctx, "pub-789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
}

// TestGetRootTreeWithNilIdentity ensures GetRootTree works without identity.
func TestGetRootTreeWithNilIdentityValue(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":1,"name":"Root","parent_type_id":null}]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	ctx := WithIdentity(context.Background(), nil)
	_, err := client.GetRootTree(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNilOrEmpty(t, received, "X-User-ID")
}

// TestListObjectsWithNilContext ensures ListObjects works without identity.
func TestGetObjectTypeByIDForwardsAllIdentityHeaders(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "full-id")
	hdr.Set("X-User-Email", "full@example.com")
	hdr.Set("X-User-Roles", "admin,user")
	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.GetObjectTypeByID(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "full-id")
	assertHeader(t, received, "X-User-Email", "full@example.com")
	assertHeader(t, received, "X-User-Roles", "admin,user")
}

// TestGetObjectByPublicIDForwardsAllIdentityHeaders ensures GetObjectByPublicID forwards all identity headers.
func TestGetObjectByPublicIDForwardsAllIdentityHeaders(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "full-id")
	hdr.Set("X-User-Email", "full@example.com")
	hdr.Set("X-User-Roles", "admin,user")
	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.GetObjectByPublicID(ctx, "pub-full")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "full-id")
	assertHeader(t, received, "X-User-Email", "full@example.com")
	assertHeader(t, received, "X-User-Roles", "admin,user")
}

// TestGetObjectTypeByNameForwardsAllIdentityHeaders ensures GetObjectTypeByName forwards all identity headers.
func TestGetObjectTypeByNameForwardsAllIdentityHeaders(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "full-id")
	hdr.Set("X-User-Email", "full@example.com")
	hdr.Set("X-User-Roles", "admin,user")
	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.GetObjectTypeByName(ctx, "document")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "full-id")
	assertHeader(t, received, "X-User-Email", "full@example.com")
	assertHeader(t, received, "X-User-Roles", "admin,user")
}

// TestGetObjectTypeTreeForwardsAllIdentityHeaders ensures GetObjectTypeTree forwards all identity headers.
func TestGetObjectTypeTreeForwardsAllIdentityHeaders(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"id":1}}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "full-id")
	hdr.Set("X-User-Email", "full@example.com")
	hdr.Set("X-User-Roles", "admin,user")
	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.GetObjectTypeTree(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "full-id")
	assertHeader(t, received, "X-User-Email", "full@example.com")
	assertHeader(t, received, "X-User-Roles", "admin,user")
}

// TestGetRootTreeForwardsAllIdentityHeaders ensures GetRootTree forwards all identity headers.
func TestGetRootTreeForwardsAllIdentityHeaders(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":1,"name":"Root","parent_type_id":null}]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "full-id")
	hdr.Set("X-User-Email", "full@example.com")
	hdr.Set("X-User-Roles", "admin,user")
	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.GetRootTree(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "full-id")
	assertHeader(t, received, "X-User-Email", "full@example.com")
	assertHeader(t, received, "X-User-Roles", "admin,user")
}

// TestListObjectTypesForwardsAllIdentityHeaders ensures ListObjectTypes forwards all identity headers.
func TestListObjectTypesForwardsAllIdentityHeaders(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewObjectsClient(server.URL, 0)
	hdr := http.Header{}
	hdr.Set("X-User-ID", "full-id")
	hdr.Set("X-User-Email", "full@example.com")
	hdr.Set("X-User-Roles", "admin,user")
	ctx := WithIdentity(context.Background(), hdr)
	_, err := client.ListObjectTypes(ctx, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHeader(t, received, "X-User-ID", "full-id")
	assertHeader(t, received, "X-User-Email", "full@example.com")
	assertHeader(t, received, "X-User-Roles", "admin,user")
}

// TestForwardHeadersUnexported verifies forwardHeaders is not exported (compile-time check).
func TestForwardHeadersUnexported(t *testing.T) {
	// This test exists to document the design decision that forwardHeaders is unexported.
	// If someone tries to import client.forwardHeaders, it will fail at compile time.
	// The allow-list is intentionally private — only httpGet should iterate over it.
	if len(forwardHeaders) != 3 {
		t.Fatalf("forwardHeaders should have exactly 3 entries")
	}
	for _, h := range forwardHeaders {
		if !strings.HasPrefix(h, "X-User-") {
			t.Errorf("all forward headers must start with 'X-User-', got %q", h)
		}
	}
}
