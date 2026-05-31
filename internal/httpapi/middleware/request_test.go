package middleware

import (
	"errors"
	"strings"
	"testing"
)

func TestNewRequestIDFallbackRemainsUnique(t *testing.T) {
	originalReader := requestRandomRead
	originalCounter := requestFallbackCounter.Load()
	requestRandomRead = func([]byte) (int, error) {
		return 0, errors.New("entropy unavailable")
	}
	requestFallbackCounter.Store(0)
	t.Cleanup(func() {
		requestRandomRead = originalReader
		requestFallbackCounter.Store(originalCounter)
	})

	first := newRequestID("req")
	second := newRequestID("req")

	if !strings.HasPrefix(first, "req_") || !strings.HasPrefix(second, "req_") {
		t.Fatalf("expected request ids to keep prefix, got %q and %q", first, second)
	}
	if first == "req_fallback" || second == "req_fallback" {
		t.Fatal("expected fallback request id to avoid a constant value")
	}
	if first == second {
		t.Fatalf("expected distinct fallback request ids, got %q", first)
	}
}
