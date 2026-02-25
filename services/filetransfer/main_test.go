package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFiletransferRateLimit(t *testing.T) {
	fileLimiter = newIPRateLimiter(1, time.Hour)

	h := withRequestTrace("health", 0, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "10.0.0.10:1234"
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/health", nil)
	req2.RemoteAddr = "10.0.0.10:1234"
	rec2 := httptest.NewRecorder()
	h(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d body=%s", rec2.Code, rec2.Body.String())
	}
}

func TestFiletransferUploadRequiresMultipart(t *testing.T) {
	fileLimiter = newIPRateLimiter(1000, time.Hour)

	svc := NewFileTransferService(t.TempDir())
	h := withRequestTrace("upload", defaultUploadLimit, svc.UploadHandler)

	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("hi"))
	req.Header.Set("Content-Type", "text/plain")
	req.RemoteAddr = "10.0.0.11:1"
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestFiletransferDownloadValidatesUUID(t *testing.T) {
	fileLimiter = newIPRateLimiter(1000, time.Hour)

	svc := NewFileTransferService(t.TempDir())
	h := withRequestTrace("download", 0, svc.DownloadHandler)

	req := httptest.NewRequest(http.MethodGet, "/download?id=not-a-uuid", nil)
	req.RemoteAddr = "10.0.0.12:1"
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}
