package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthRateLimit(t *testing.T) {
	authLimiter = newIPRateLimiter(2, time.Hour)

	h := withRequestTrace("login", defaultAuthBodyLimit, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"username":"a","password":"b"}`))
		req.RemoteAddr = "10.0.0.1:1234"
		rec := httptest.NewRecorder()
		h(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"username":"a","password":"b"}`))
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthRegisterValidation(t *testing.T) {
	svc, err := NewAuthService(":memory:")
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	h := withRequestTrace("register", defaultAuthBodyLimit, svc.RegisterHandler)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"username":"!!","full_name":"","password":"short","role":"x"}`))
	req.RemoteAddr = "10.0.0.2:9999"
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthBodyLimit(t *testing.T) {
	authLimiter = newIPRateLimiter(1000, time.Hour)
	svc, err := NewAuthService(":memory:")
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}

	// Exceed the MaxBytesReader limit; handler should reject with 4xx.
	tooBig := bytes.Repeat([]byte("a"), int(defaultAuthBodyLimit)+1)
	h := withRequestTrace("login", defaultAuthBodyLimit, svc.LoginHandler)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(tooBig))
	req.RemoteAddr = "10.0.0.3:1"
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code < 400 || rec.Code >= 500 {
		t.Fatalf("expected 4xx, got %d body=%s", rec.Code, rec.Body.String())
	}
}
