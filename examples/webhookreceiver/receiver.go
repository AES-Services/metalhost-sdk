// Package webhookreceiver demonstrates bounded V2 verification and durable handoff.
// It is not a standalone server: supply a durable Inbox and serve behind HTTPS.
package webhookreceiver

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/AES-Services/metalhost-sdk/metalhost"
)

// Inbox must atomically persist the verified body with a unique DeliveryID.
// Duplicate IDs are successful no-ops, not errors. A worker processes persisted
// bodies and deduplicates individual event IDs where grouped payloads overlap.
// Scope the inbox to this subscription; never use an unverified header as a key.
type Inbox interface {
	PutOnce(context.Context, string, []byte) error
}

func Handler(secret string, inbox Inbox) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		if inbox == nil {
			http.Error(w, "receiver unavailable", http.StatusServiceUnavailable)
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err != nil {
			http.Error(w, "invalid or oversized body", http.StatusBadRequest)
			return
		}
		identity, err := metalhost.VerifyWebhook(r.Header, body, secret, time.Now())
		if err != nil {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
		// Acknowledge only after durable storage. Do not log the raw body or secret.
		if err := inbox.PutOnce(r.Context(), identity.DeliveryID, body); err != nil {
			http.Error(w, "receiver unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
}
