package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/domain"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/observability"
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

func TestIdempotencyConcurrentSameKeyRunsHandlerOnce(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	store := NewIdempotencyStoreWithTTL(time.Minute)
	metrics := observability.NewMetrics("bankport_idempotency_test")
	var handlerCalls atomic.Int32
	handlerStarted := make(chan struct{})
	releaseHandler := make(chan struct{})

	router := gin.New()
	router.Use(RequestIdentity())
	router.Use(func(c *gin.Context) {
		c.Set(PartnerKey, domain.Partner{ID: "partner_test", DeveloperAppID: "app_test"})
		c.Next()
	})
	router.POST("/financial-writes", Idempotency(store, metrics), func(c *gin.Context) {
		call := handlerCalls.Add(1)
		if call == 1 {
			close(handlerStarted)
			<-releaseHandler
		}
		c.JSON(http.StatusAccepted, gin.H{"accepted": true, "call": call})
	})

	firstDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		firstDone <- performIdempotencyRequest(router, "same-key", `{"amount_cents":100}`)
	}()
	<-handlerStarted

	secondDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		secondDone <- performIdempotencyRequest(router, "same-key", `{"amount_cents":100}`)
	}()

	close(releaseHandler)
	first := <-firstDone
	second := <-secondDone

	if first.Code != http.StatusAccepted {
		t.Fatalf("expected first request 202, got %d: %s", first.Code, first.Body.String())
	}
	if second.Code != http.StatusAccepted {
		t.Fatalf("expected concurrent replay 202, got %d: %s", second.Code, second.Body.String())
	}
	if second.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatalf("expected second request to replay cached response, got header %q", second.Header().Get("Idempotency-Replayed"))
	}
	if first.Body.String() != second.Body.String() {
		t.Fatalf("expected replay body to match first response\nfirst: %s\nsecond: %s", first.Body.String(), second.Body.String())
	}
	if handlerCalls.Load() != 1 {
		t.Fatalf("expected handler to run once, ran %d times", handlerCalls.Load())
	}
}

func performIdempotencyRequest(router *gin.Engine, key, body string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/financial-writes", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", key)
	router.ServeHTTP(recorder, request)
	return recorder
}
