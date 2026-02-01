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
)

var testSecret = []byte("test-secret")

func signPayload(payload, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestHandler_ValidPayload(t *testing.T) {
	payload, err := os.ReadFile("../../testdata/registry_package_published.json")
	if err != nil {
		t.Fatalf("failed to read test payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandler_ValidSignature(t *testing.T) {
	payload := []byte(`{"action": "published"}`)

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandler_InvalidSignature(t *testing.T) {
	payload := []byte(`{"action": "published"}`)

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", "sha256=invalidsignature00000000000000000000000000000000000000000000000000")
	rec := httptest.NewRecorder()

	Handler(testSecret).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestHandler_MissingSignature(t *testing.T) {
	payload := []byte(`{"action": "published"}`)

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	Handler(testSecret).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestHandler_InvalidJSON(t *testing.T) {
	payload := []byte("not valid json")

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/version/1", nil)
			rec := httptest.NewRecorder()

			Handler(testSecret).ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d for %s, got %d", http.StatusMethodNotAllowed, method, rec.Code)
			}
		})
	}
}

func TestHandler_EmptyBody(t *testing.T) {
	payload := []byte("")

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandler_MissingFields(t *testing.T) {
	payload := []byte(`{"action": "published"}`)

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandler_PingEvent(t *testing.T) {
	payload := []byte(`{"zen": "Design for failure.", "hook_id": 123}`)

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", signPayload(payload, testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandler_FormEncodedRejected(t *testing.T) {
	jsonPayload := `{"action":"published","registry_package":{"name":"test","ecosystem":"docker","package_version":{"version":"1.0.0","package_url":"test","container_metadata":{"tag":{"name":"1.0.0","digest":"sha256:abc"}}}},"repository":{"full_name":"test/test"},"sender":{"login":"test"}}`

	// Create form-encoded body
	formData := url.Values{}
	formData.Set("payload", jsonPayload)
	formBody := formData.Encode()

	req := httptest.NewRequest(http.MethodPost, "/version/1", strings.NewReader(formBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Hub-Signature-256", signPayload([]byte(formBody), testSecret))
	rec := httptest.NewRecorder()

	Handler(testSecret).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected status %d, got %d: %s", http.StatusUnsupportedMediaType, rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "application/json") {
		t.Errorf("error message should mention application/json, got: %s", body)
	}
}
