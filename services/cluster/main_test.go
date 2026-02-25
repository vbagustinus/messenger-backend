package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClusterRateLimit(t *testing.T) {
	clusterLimiter = newIPRateLimiter(1, time.Hour)

	h := withRequestTrace("join", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/join?node_id=n1&addr=127.0.0.1:10001", nil)
	req.RemoteAddr = "10.0.0.30:1"
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/join?node_id=n1&addr=127.0.0.1:10001", nil)
	req2.RemoteAddr = "10.0.0.30:1"
	rec2 := httptest.NewRecorder()
	h(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d body=%s", rec2.Code, rec2.Body.String())
	}
}

func TestClusterJoinValidation(t *testing.T) {
	clusterLimiter = newIPRateLimiter(1000, time.Hour)

	node := NewRaftNode("n1", "./raft-data", ":10001")
	h := withRequestTrace("join", node.JoinHandler)

	req := httptest.NewRequest(http.MethodGet, "/join?node_id=&addr=", nil)
	req.RemoteAddr = "10.0.0.31:1"
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}
