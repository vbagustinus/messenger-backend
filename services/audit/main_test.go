package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuditRateLimit(t *testing.T) {
	auditLimiter = newIPRateLimiter(1, time.Hour)

	h := withRequestTrace("log", defaultAuditBodyLimit, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/log", bytes.NewBufferString(`{}`))
	req.RemoteAddr = "10.0.0.20:1"
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/log", bytes.NewBufferString(`{}`))
	req2.RemoteAddr = "10.0.0.20:1"
	rec2 := httptest.NewRecorder()
	h(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d body=%s", rec2.Code, rec2.Body.String())
	}
}

func TestAuditValidation(t *testing.T) {
	auditLimiter = newIPRateLimiter(1000, time.Hour)

	svc := NewAuditService(t.TempDir() + "/audit.log")
	h := withRequestTrace("log", defaultAuditBodyLimit, svc.LogHandler)

	// Invalid action (fails regex)
	body := `{"actor_id":"u1","action":"$bad","target_resource":"x","details":""}`
	req := httptest.NewRequest(http.MethodPost, "/log", bytes.NewBufferString(body))
	req.RemoteAddr = "10.0.0.21:1"
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuditBodyLimit(t *testing.T) {
	auditLimiter = newIPRateLimiter(1000, time.Hour)

	svc := NewAuditService(t.TempDir() + "/audit.log")
	h := withRequestTrace("log", defaultAuditBodyLimit, svc.LogHandler)

	tooBig := bytes.Repeat([]byte("a"), int(defaultAuditBodyLimit)+1)
	req := httptest.NewRequest(http.MethodPost, "/log", bytes.NewReader(tooBig))
	req.RemoteAddr = "10.0.0.22:1"
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code < 400 || rec.Code >= 500 {
		t.Fatalf("expected 4xx, got %d body=%s", rec.Code, rec.Body.String())
	}
}
