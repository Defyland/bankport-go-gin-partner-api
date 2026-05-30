package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/observability"
)

type IdempotencyStore struct {
	mu      sync.RWMutex
	ttl     time.Duration
	clock   func() time.Time
	records map[string]IdempotencyRecord
}

type IdempotencyRecord struct {
	RequestHash string
	Status      int
	Header      http.Header
	Body        []byte
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

func NewIdempotencyStore() *IdempotencyStore {
	return NewIdempotencyStoreWithTTL(24 * time.Hour)
}

func NewIdempotencyStoreWithTTL(ttl time.Duration) *IdempotencyStore {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &IdempotencyStore{
		ttl:     ttl,
		clock:   func() time.Time { return time.Now().UTC() },
		records: make(map[string]IdempotencyRecord),
	}
}

func (s *IdempotencyStore) Get(key string) (IdempotencyRecord, bool) {
	now := s.clock()
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.records[key]
	if !ok {
		return IdempotencyRecord{}, false
	}
	if !record.ExpiresAt.IsZero() && !now.Before(record.ExpiresAt) {
		delete(s.records, key)
		return IdempotencyRecord{}, false
	}
	return record, ok
}

func (s *IdempotencyStore) Put(key string, record IdempotencyRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}
	if record.ExpiresAt.IsZero() {
		record.ExpiresAt = record.CreatedAt.Add(s.ttl)
	}
	s.pruneExpiredLocked(now)
	s.records[key] = record
}

func (s *IdempotencyStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.records)
}

func Idempotency(store *IdempotencyStore, metrics *observability.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		partner, ok := Partner(c)
		if !ok {
			Abort(c, http.StatusUnauthorized, "authentication_required", "Authentication context is missing.", nil)
			return
		}

		idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
		if idempotencyKey == "" {
			Abort(c, http.StatusBadRequest, "idempotency_key_required", "Financial write requests require an Idempotency-Key header.", nil)
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			Abort(c, http.StatusBadRequest, "invalid_request_body", "The request body could not be read.", nil)
			return
		}
		_ = c.Request.Body.Close()
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		routeName := route(c)
		recordKey := partner.ID + "|" + routeName + "|" + idempotencyKey
		requestHash := requestHash(c.Request.Method, routeName, body)
		if record, found := store.Get(recordKey); found {
			if record.RequestHash != requestHash {
				metrics.IdempotencyConflicts.WithLabelValues(partner.ID, routeName).Inc()
				Abort(c, http.StatusConflict, "idempotency_conflict", "The Idempotency-Key was already used with a different request body.", nil)
				return
			}
			for key, values := range record.Header {
				for _, value := range values {
					c.Writer.Header().Add(key, value)
				}
			}
			c.Header("Idempotency-Replayed", "true")
			metrics.IdempotencyReplays.WithLabelValues(partner.ID, routeName).Inc()
			c.Data(record.Status, record.Header.Get("Content-Type"), record.Body)
			c.Abort()
			return
		}

		writer := &captureWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = writer
		c.Next()

		status := statusFromWriter(c)
		if status < http.StatusInternalServerError {
			store.Put(recordKey, IdempotencyRecord{
				RequestHash: requestHash,
				Status:      status,
				Header:      cloneHeader(c.Writer.Header()),
				Body:        append([]byte(nil), writer.body.Bytes()...),
				CreatedAt:   time.Now().UTC(),
			})
		}
	}
}

func (s *IdempotencyStore) pruneExpiredLocked(now time.Time) {
	for key, record := range s.records {
		if !record.ExpiresAt.IsZero() && !now.Before(record.ExpiresAt) {
			delete(s.records, key)
		}
	}
}

type captureWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *captureWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *captureWriter) WriteString(data string) (int, error) {
	w.body.WriteString(data)
	return w.ResponseWriter.WriteString(data)
}

func requestHash(method, route string, body []byte) string {
	sum := sha256.Sum256(append([]byte(method+"|"+route+"|"), body...))
	return hex.EncodeToString(sum[:])
}

func cloneHeader(header http.Header) http.Header {
	cloned := make(http.Header, len(header))
	for key, values := range header {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}
