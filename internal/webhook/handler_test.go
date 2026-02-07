package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/jc/steakpie/internal/config"
	"github.com/jc/steakpie/internal/executor"
)

var testSecret = []byte("test-secret")

var testConfig = config.Config{
	"test-package": {
		{Cmd: "echo test"},
		{Cmd: "docker compose up"},
	},
	"jamiec": {
		{Cmd: "docker compose down"},
		{Cmd: "docker compose up"},
	},
	"hello-world": {
		{Cmd: "echo hello"},
		{Cmd: "docker pull hello-world"},
	},
}

// noopRunner is a no-op runner for handler tests.
type noopRunner struct{}

func (noopRunner) Run(cmd string) (string, error) { return "", nil }

var testRunner executor.Runner = noopRunner{}

func signPayload(payload, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// createTestStore creates an in-memory EventStore for testing
func createTestStore(t *testing.T) *EventStore {
	store, err := NewEventStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestHandler_ValidPayload(t *testing.T) {
	store := createTestStore(t)

	payload, err := os.ReadFile("../../testdata/registry_package_published.json")
	if err != nil {
		t.Fatalf("failed to read test payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret, testConfig, store, testRunner).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandler_ValidSignature(t *testing.T) {
	store := createTestStore(t)

	payload := []byte(`{"action": "published"}`)

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret, testConfig, store, testRunner).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandler_InvalidSignature(t *testing.T) {
	store := createTestStore(t)

	payload := []byte(`{"action": "published"}`)

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", "sha256=invalidsignature00000000000000000000000000000000000000000000000000")
	rec := httptest.NewRecorder()

	Handler(testSecret, testConfig, store, testRunner).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestHandler_MissingSignature(t *testing.T) {
	store := createTestStore(t)

	payload := []byte(`{"action": "published"}`)

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	Handler(testSecret, testConfig, store, testRunner).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestHandler_InvalidJSON(t *testing.T) {
	store := createTestStore(t)

	payload := []byte("not valid json")

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret, testConfig, store, testRunner).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	store := createTestStore(t)

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/version/1", nil)
			rec := httptest.NewRecorder()

			Handler(testSecret, testConfig, store, testRunner).ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d for %s, got %d", http.StatusMethodNotAllowed, method, rec.Code)
			}
		})
	}
}

func TestHandler_EmptyBody(t *testing.T) {
	store := createTestStore(t)

	payload := []byte("")

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret, testConfig, store, testRunner).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandler_MissingFields(t *testing.T) {
	store := createTestStore(t)

	payload := []byte(`{"action": "published"}`)

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret, testConfig, store, testRunner).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandler_PingEvent(t *testing.T) {
	store := createTestStore(t)

	payload := []byte(`{"zen": "Design for failure.", "hook_id": 123}`)

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret, testConfig, store, testRunner).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandler_FormEncodedRejected(t *testing.T) {
	store := createTestStore(t)

	jsonPayload := `{"action":"published","registry_package":{"name":"test","ecosystem":"docker","package_version":{"version":"1.0.0","package_url":"test","container_metadata":{"tag":{"name":"1.0.0","digest":"sha256:abc"}}}},"repository":{"full_name":"test/test"},"sender":{"login":"test"}}`

	// Create form-encoded body
	formData := url.Values{}
	formData.Set("payload", jsonPayload)
	formBody := formData.Encode()

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(formBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Hub-Signature-256", signPayload([]byte(formBody), testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret, testConfig, store, testRunner).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected status %d, got %d: %s", http.StatusUnsupportedMediaType, rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "application/json") {
		t.Errorf("error message should mention application/json, got: %s", body)
	}
}

func TestHandler_ConfiguredPackage(t *testing.T) {
	store := createTestStore(t)

	payload, err := os.ReadFile("../../testdata/registry_package_published.json")
	if err != nil {
		t.Fatalf("failed to read test payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret, testConfig, store, testRunner).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandler_UnconfiguredPackage(t *testing.T) {
	store := createTestStore(t)

	// Create a payload with an unconfigured package name
	payload := []byte(`{
		"action": "published",
		"registry_package": {
			"name": "unconfigured-package",
			"package_version": {
				"version": "1.0.0"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret, testConfig, store, testRunner).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandler_Idempotency_DuplicateEvent(t *testing.T) {
	// Use a temp database file for this test
	tmpDB := t.TempDir() + "/test_db.sqlite"
	store, err := NewEventStore(tmpDB)
	if err != nil {
		t.Fatalf("failed to create event store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	payload, err := os.ReadFile("../../testdata/registry_package_published.json")
	if err != nil {
		t.Fatalf("failed to read test payload: %v", err)
	}

	deliveryID := "test-delivery-001"

	// Create handler
	handler := Handler(testSecret, testConfig, store, testRunner)

	// First request
	req1 := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	req1.Header.Set("X-GitHub-Delivery", deliveryID)
	rec1 := httptest.NewRecorder()

	handler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Errorf("First request: expected status %d, got %d", http.StatusOK, rec1.Code)
	}

	// Second request with same delivery ID (duplicate)
	req2 := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	req2.Header.Set("X-GitHub-Delivery", deliveryID)
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("Second request: expected status %d, got %d", http.StatusOK, rec2.Code)
	}

	// Both should return 200, but second should be logged as duplicate
}

func TestHandler_Idempotency_DifferentEvents(t *testing.T) {
	tmpDB := t.TempDir() + "/test_db.sqlite"
	store, err := NewEventStore(tmpDB)
	if err != nil {
		t.Fatalf("failed to create event store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	payload, err := os.ReadFile("../../testdata/registry_package_published.json")
	if err != nil {
		t.Fatalf("failed to read test payload: %v", err)
	}

	handler := Handler(testSecret, testConfig, store, testRunner)

	deliveryIDs := []string{"delivery-001", "delivery-002", "delivery-003"}

	for _, deliveryID := range deliveryIDs {
		req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
		req.Header.Set("X-GitHub-Delivery", deliveryID)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %s: expected status %d, got %d", deliveryID, http.StatusOK, rec.Code)
		}
	}
}

func TestHandler_Idempotency_MissingDeliveryID(t *testing.T) {
	tmpDB := t.TempDir() + "/test_db.sqlite"
	store, err := NewEventStore(tmpDB)
	if err != nil {
		t.Fatalf("failed to create event store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	payload, err := os.ReadFile("../../testdata/registry_package_published.json")
	if err != nil {
		t.Fatalf("failed to read test payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	// Intentionally NOT setting X-GitHub-Delivery header
	rec := httptest.NewRecorder()

	Handler(testSecret, testConfig, store, testRunner).ServeHTTP(rec, req)

	// Should still return 200 (backwards compatibility)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandler_Idempotency_PingEventsNotDeduplicated(t *testing.T) {
	tmpDB := t.TempDir() + "/test_db.sqlite"
	store, err := NewEventStore(tmpDB)
	if err != nil {
		t.Fatalf("failed to create event store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	payload := []byte(`{"zen": "Design for failure.", "hook_id": 123}`)
	deliveryID := "ping-delivery-001"

	handler := Handler(testSecret, testConfig, store, testRunner)

	// Send ping event twice with same delivery ID
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
		req.Header.Set("X-GitHub-Delivery", deliveryID)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Ping request %d: expected status %d, got %d", i+1, http.StatusOK, rec.Code)
		}
	}
}
