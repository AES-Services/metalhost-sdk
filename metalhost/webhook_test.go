package metalhost

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestVerifyWebhookRotationReplayAndTampering(t *testing.T) {
	now := time.Unix(1789920000, 0)
	secret, old := strings.Repeat("n", 40), strings.Repeat("o", 40)
	body := []byte(`{"event_ids":["stable-event"]}`)
	headers := make(http.Header)
	headers.Set("X-Metalhost-Delivery", "group/stable-delivery")
	headers.Set("X-Metalhost-Attempt", "00000000-0000-4000-8000-000000000001")
	timestamp := strconv.FormatInt(now.Unix(), 10)
	sign := func(key string) string {
		mac := hmac.New(sha256.New, []byte(key))
		_, _ = mac.Write(append([]byte(timestamp+".group/stable-delivery.00000000-0000-4000-8000-000000000001."), body...))
		return hex.EncodeToString(mac.Sum(nil))
	}
	headers.Set("X-Metalhost-Signature-V2", "t="+timestamp+",v2="+sign(secret)+",v2="+sign(old))
	for _, key := range []string{secret, old} {
		identity, err := VerifyWebhook(headers, body, key, now)
		if err != nil || identity.DeliveryID != "group/stable-delivery" || identity.AttemptID != "00000000-0000-4000-8000-000000000001" || !identity.SignedAt.Equal(now) {
			t.Fatalf("overlap verification: %+v %v", identity, err)
		}
	}
	for _, delta := range []time.Duration{6 * time.Minute, -6 * time.Minute} {
		if _, err := VerifyWebhook(headers, body, secret, now.Add(delta)); err == nil {
			t.Fatal("replay window ignored")
		}
	}
	for _, field := range []string{"X-Metalhost-Delivery", "X-Metalhost-Attempt", "X-Metalhost-Signature-V2"} {
		changed := headers.Clone()
		changed.Set(field, "tampered")
		if _, err := VerifyWebhook(changed, body, secret, now); err == nil {
			t.Fatalf("accepted changed %s", field)
		}
		changed = headers.Clone()
		changed.Add(field, headers.Get(field))
		if _, err := VerifyWebhook(changed, body, secret, now); err == nil {
			t.Fatalf("accepted duplicate %s", field)
		}
	}
	for _, raw := range [][]byte{[]byte(`{ "event_ids":["stable-event"]}`), make([]byte, (1<<20)+1)} {
		if _, err := VerifyWebhook(headers, raw, secret, now); err == nil {
			t.Fatal("accepted altered/oversized body")
		}
	}
	headers.Del("X-Metalhost-Signature-V2")
	headers.Set("X-Metalhost-Signature", "sha256="+sign(secret))
	if _, err := VerifyWebhook(headers, body, secret, now); err == nil {
		t.Fatal("downgraded to legacy signature")
	}
}
