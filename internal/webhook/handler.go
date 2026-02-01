package webhook

import (
	"encoding/json"
	"io"
	"net/http"
)

// Handler returns an HTTP handler for registry_package webhook events.
// The secret is used to verify the webhook signature.
func Handler(secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		signature := r.Header.Get("X-Hub-Signature-256")
		if !VerifySignature(body, signature, secret) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		var event RegistryPackageEvent
		if err := json.Unmarshal(body, &event); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
