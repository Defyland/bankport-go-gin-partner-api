package middleware

import (
	"net/http"
	"testing"
	"time"
)

func TestIdempotencyStoreExpiresRecords(t *testing.T) {
	now := time.Date(2026, 5, 30, 12, 0, 0, 0, time.UTC)
	store := NewIdempotencyStoreWithTTL(time.Minute)
	store.clock = func() time.Time { return now }

	store.Put("partner|route|key", IdempotencyRecord{
		RequestHash: "hash",
		Status:      http.StatusAccepted,
		Body:        []byte(`{"ok":true}`),
	})
	if _, ok := store.Get("partner|route|key"); !ok {
		t.Fatal("expected fresh idempotency record")
	}

	now = now.Add(time.Minute)
	if _, ok := store.Get("partner|route|key"); ok {
		t.Fatal("expected expired idempotency record to be removed")
	}
	if store.Len() != 0 {
		t.Fatalf("expected store cleanup, got %d records", store.Len())
	}
}
