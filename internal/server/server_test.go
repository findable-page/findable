package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nadimtuhin/seoscore-cli/internal/audit"
)

func newTestMux() *http.ServeMux {
	return NewMux(&http.Client{}, false)
}

func fixturePage(t *testing.T) (*http.ServeMux, string) {
	t.Helper()
	site := httptest.NewServer(http.FileServer(http.Dir("../../testdata/site")))
	t.Cleanup(site.Close)
	return NewMux(site.Client(), false), site.URL + "/01-thin-poor.html"
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

func TestServerScanLivePage(t *testing.T) {
	mux, page := fixturePage(t)
	req := newRequest(http.MethodPost, "/api/scan", `{"url":"`+page+`"}`)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
	var res audit.Scan
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res.SEO.Score == nil || *res.SEO.Score <= 0 {
		t.Errorf("expected valid SEO score, got %v", res.SEO.Score)
	}
}

func TestServerScanFormats(t *testing.T) {
	for _, tc := range []struct {
		format, contentType, prefix string
	}{
		{"sarif", "application/sarif+json", `{"$schema":"https://docs.oasis-open.org/sarif/`},
		{"csv", "text/csv; charset=utf-8", "\xef\xbb\xbfSEO/AEO/GEO Report,127.0.0.1:"},
	} {
		mux, page := fixturePage(t)
		rec := httptest.NewRecorder()
		body := `{"url":"` + page + `","format":"` + tc.format + `"}`
		mux.ServeHTTP(rec, newRequest(http.MethodPost, "/api/scan", body))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d: %s", tc.format, rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("Content-Type"); got != tc.contentType {
			t.Errorf("%s: Content-Type = %q", tc.format, got)
		}
		if !strings.HasPrefix(rec.Body.String(), tc.prefix) {
			t.Errorf("%s: body starts %q", tc.format, rec.Body.String()[:min(80, rec.Body.Len())])
		}
	}
}

func TestServerScanRejectsUnknownFormat(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestMux().ServeHTTP(rec, newRequest(http.MethodPost, "/api/scan", `{"url":"../../testdata/site/01-thin-poor.html","format":"xml"}`))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestServerScanAddsSiteChecksOnlyWhenAsked(t *testing.T) {
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nDisallow:\n"))
			return
		}
		_, _ = w.Write([]byte("<html><head><title>x</title></head><body></body></html>"))
	}))
	t.Cleanup(site.Close)
	mux := NewMux(site.Client(), false)
	for _, tc := range []struct {
		body     string
		wantSite bool
	}{
		{`{"url":"` + site.URL + `/"}`, false},
		{`{"url":"` + site.URL + `/","site":true}`, true},
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, newRequest(http.MethodPost, "/api/scan", tc.body))
		var res struct {
			Site *struct {
				Total int `json:"total"`
			} `json:"site"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("%s: %v %s", tc.body, err, rec.Body.String())
		}
		if (res.Site != nil) != tc.wantSite {
			t.Errorf("%s: site present = %v, want %v", tc.body, res.Site != nil, tc.wantSite)
		}
	}
}

func TestGuardedServerRejectsLocalFilesAndPrivateAddresses(t *testing.T) {
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html></html>"))
	}))
	t.Cleanup(site.Close)
	handler := NewServer(":0", false).Handler
	for _, target := range []string{"../../testdata/site/01-thin-poor.html", "file:///etc/passwd", "ftp://example.com/", site.URL} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, newRequest(http.MethodPost, "/api/scan", `{"url":"`+target+`"}`))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400 (%s)", target, rec.Code, rec.Body.String())
		}
	}
}
