package metalhost

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// WebhookIdentity is returned only after signature and replay-window validation.
// Deduplicate DeliveryID durably. AttemptID changes on each delivery retry.
type WebhookIdentity struct {
	DeliveryID string
	AttemptID  string
	SignedAt   time.Time
}

var webhookAttempt = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
var webhookDelivery = regexp.MustCompile(`^[a-zA-Z0-9_/-]{1,512}$`)

// VerifyWebhook verifies the timestamped V2 signature over the exact raw body.
// Bound request bodies before reading them (maximum 1 MiB here), pass a trusted
// clock, and persist deduplication before acknowledging the delivery. Signatures
// alone do not prevent repeated delivery within the five-minute replay window.
// Either overlapping secret is accepted if its signature is present; the legacy
// untimestamped signature header is deliberately never used as a fallback.
func VerifyWebhook(headers http.Header, body []byte, secret string, now time.Time) (WebhookIdentity, error) {
	invalid := func() (WebhookIdentity, error) {
		return WebhookIdentity{}, errors.New("invalid or expired webhook signature")
	}
	if len(body) > 1<<20 || len(secret) < 32 || len(secret) > 256 || now.IsZero() {
		return invalid()
	}
	one := func(name string) string {
		values := headers.Values(name)
		if len(values) != 1 {
			return ""
		}
		return values[0]
	}
	delivery, attempt, signature := one("X-Metalhost-Delivery"), one("X-Metalhost-Attempt"), one("X-Metalhost-Signature-V2")
	if !webhookDelivery.MatchString(delivery) || !webhookAttempt.MatchString(attempt) || len(signature) > 1024 {
		return invalid()
	}
	fields := strings.Split(signature, ",")
	if len(fields) < 2 || len(fields) > 3 || !strings.HasPrefix(fields[0], "t=") {
		return invalid()
	}
	timestamp := strings.TrimPrefix(fields[0], "t=")
	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil || seconds <= 0 || strconv.FormatInt(seconds, 10) != timestamp {
		return invalid()
	}
	signedAt := time.Unix(seconds, 0)
	age := now.Sub(signedAt)
	if age > 5*time.Minute || age < -5*time.Minute {
		return invalid()
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "." + delivery + "." + attempt + "."))
	_, _ = mac.Write(body)
	expected := mac.Sum(nil)
	matched := false
	for _, field := range fields[1:] {
		if !strings.HasPrefix(field, "v2=") {
			return invalid()
		}
		supplied, err := hex.DecodeString(strings.TrimPrefix(field, "v2="))
		if err != nil || len(supplied) != sha256.Size {
			return invalid()
		}
		valid := hmac.Equal(expected, supplied)
		matched = matched || valid
	}
	if !matched {
		return invalid()
	}
	return WebhookIdentity{DeliveryID: delivery, AttemptID: attempt, SignedAt: signedAt}, nil
}
