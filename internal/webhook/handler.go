package webhook

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

// Handler returns an HTTP handler for registry_package webhook events.
// The secret is used to verify the webhook signature.
func Handler(secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request from %s", r.Method, r.RemoteAddr)

		if r.Method != http.MethodPost {
			log.Printf("Method not allowed: %s", r.Method)
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check Content-Type - only accept JSON
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
			log.Printf("Rejected form-encoded webhook - application/json required")
			errorMsg := "Form-encoded webhooks are not supported.\n\n" +
				"Please configure your GitHub webhook to use application/json:\n" +
				"1. Go to your repository settings\n" +
				"2. Navigate to Webhooks\n" +
				"3. Edit the webhook\n" +
				"4. Change 'Content type' to 'application/json'\n" +
				"5. Save the webhook"
			http.Error(w, errorMsg, http.StatusUnsupportedMediaType)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read request body: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		signature := r.Header.Get("X-Hub-Signature-256")
		if signature == "" {
			log.Printf("Missing X-Hub-Signature-256 header")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if !VerifySignature(body, signature, secret) {
			log.Printf("Signature verification failed - received: %s", signature)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Check if this is a ping event
		var rawEvent map[string]interface{}
		if err := json.Unmarshal(body, &rawEvent); err != nil {
			log.Printf("Failed to parse JSON payload: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Handle ping events
		if zen, ok := rawEvent["zen"].(string); ok {
			log.Printf("✓ Received ping event: %s", zen)
			w.WriteHeader(http.StatusOK)
			return
		}

		var event RegistryPackageEvent
		if err := json.Unmarshal(body, &event); err != nil {
			log.Printf("Failed to parse registry_package event: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		log.Printf("✓ Successfully processed %s event for package %s (version %s)",
			event.Action,
			event.RegistryPackage.Name,
			event.RegistryPackage.PackageVersion.Version)

		w.WriteHeader(http.StatusOK)
	}
}
