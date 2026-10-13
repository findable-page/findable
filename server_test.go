package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestMux() *http.ServeMux {
	return newMux(&http.Client{}, false)
}

func newRequest(method, target, body string) *http.Request {
	return httptest.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body))
}

func TestServerHealthEndpoint(t *testing.T) {
	mux := newTestMux()
	req := newRequest(http.MethodGet, "/health", "")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	if rec.Body.String() != `{"status":"ok"}` {
		t.Errorf("unexpected body: %q", rec.Body.String())
	}
}

func TestServerScanCountEndpoint(t *testing.T) {
	mux := newTestMux()
	req := newRequest(http.MethodGet, "/api/scan-count", "")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	var data map[string]int
	if err := json.Unmarshal(rec.Body.Bytes(), &data); err != nil || data["count"] <= 0 {
		t.Errorf("expected positive count, got %v", data)
	}
}

func TestServerWebUIRender(t *testing.T) {
	mux := newTestMux()
	req := newRequest(http.MethodGet, "/", "")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "SEOScore Tools") {
		t.Errorf("expected HTML body to contain 'SEOScore Tools', got: %s", rec.Body.String()[:100])
	}
}

func TestServerCORSPreflight(t *testing.T) {
	mux := newTestMux()
	req := newRequest(http.MethodOptions, "/api/scan", "")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for OPTIONS, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *")
	}
}

func TestServerScanMissingOrEmptyURL(t *testing.T) {
	mux := newTestMux()

	// 1. Empty body
	req := newRequest(http.MethodPost, "/api/scan", "{}")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty json, got %d", rec.Code)
	}

	// 2. Invalid JSON
	req = newRequest(http.MethodPost, "/api/scan", `invalid json`)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid json, got %d", rec.Code)
	}
}

func TestServerScanLocalFile(t *testing.T) {
	mux := newTestMux()
	reqBody := `{"url":"testdata/site/01-thin-poor.html"}`
	req := newRequest(http.MethodPost, "/api/scan", reqBody)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
	var res scan
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res.SEO.Score == nil || *res.SEO.Score <= 0 {
		t.Errorf("expected valid SEO score, got %v", res.SEO.Score)
	}
}
