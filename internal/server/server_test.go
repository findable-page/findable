package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/nadimtuhin/findable/internal/audit"
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
	if !strings.Contains(rec.Body.String(), "findable") {
		t.Errorf("expected HTML body to contain 'findable', got: %s", rec.Body.String()[:100])
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

func TestServerServesTheSingleUISourceByteForByte(t *testing.T) {
	want, err := os.ReadFile("web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := newRequest(http.MethodGet, "/", "")
	req.Header.Set("Accept", "text/html")
	newTestMux().ServeHTTP(rec, req)
	if !bytes.Equal(rec.Body.Bytes(), want) {
		t.Errorf("served UI (%d bytes) differs from web/index.html (%d bytes)", rec.Body.Len(), len(want))
	}
}

func TestServerScanReportsTheCompositeScore(t *testing.T) {
	mux, page := fixturePage(t)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, newRequest(http.MethodPost, "/api/scan", `{"url":"`+page+`"}`))
	var res struct {
		Overall *int                    `json:"overall"`
		SEO     struct{ Score float64 } `json:"seo"`
		AEO     struct{ Score float64 } `json:"aeo"`
		GEO     struct{ Score float64 } `json:"geo"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil || res.Overall == nil {
		t.Fatalf("response has no overall: %v %s", err, rec.Body.String()[:min(200, rec.Body.Len())])
	}
	if want := int(res.SEO.Score*0.5 + res.AEO.Score*0.25 + res.GEO.Score*0.25 + 0.5); *res.Overall != want {
		t.Errorf("overall = %d, want %d", *res.Overall, want)
	}
}

func TestServerScanAnswers429AfterTenScansAMinuteFromOneIP(t *testing.T) {
	mux := newTestMux()
	scan := func(remoteAddr string) int {
		req := newRequest(http.MethodPost, "/api/scan", `{}`)
		req.RemoteAddr = remoteAddr
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec.Code
	}
	for range 10 {
		if code := scan("198.51.100.7:1111"); code != http.StatusBadRequest {
			t.Fatalf("scan under the limit = %d, want 400", code)
		}
	}
	if code := scan("198.51.100.7:2222"); code != http.StatusTooManyRequests {
		t.Errorf("11th scan from the same IP = %d, want 429", code)
	}
	if code := scan("198.51.100.8:1111"); code != http.StatusBadRequest {
		t.Errorf("scan from another IP = %d, want 400", code)
	}
}

func TestServerScanLimitsIPv6ClientsByTheir64(t *testing.T) {
	mux := newTestMux()
	scan := func(remoteAddr string) int {
		req := newRequest(http.MethodPost, "/api/scan", `{}`)
		req.RemoteAddr = remoteAddr
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec.Code
	}
	for i := range 10 {
		scan(fmt.Sprintf("[2001:db8:1:2::%x]:1111", i+1))
	}
	if code := scan("[2001:db8:1:2:ffff::1]:1111"); code != http.StatusTooManyRequests {
		t.Errorf("11th scan from the same /64 = %d, want 429", code)
	}
	if code := scan("[2001:db8:1:3::1]:1111"); code != http.StatusBadRequest {
		t.Errorf("scan from another /64 = %d, want 400", code)
	}
}

func TestRateLimiterResetsEachWindow(t *testing.T) {
	now := time.Unix(1000, 0)
	l := newRateLimiter(1, time.Minute, func() time.Time { return now })
	first, second := l.allow("a"), l.allow("a")
	now = now.Add(time.Minute)
	got := []bool{first, second, l.allow("a")}
	if want := []bool{true, false, true}; !slices.Equal(got, want) {
		t.Errorf("allow sequence = %v, want %v", got, want)
	}
}

func TestServerUIHasACSPHashingEveryInlineBlockAndRefusingFraming(t *testing.T) {
	req := newRequest(http.MethodGet, "/", "")
	req.Header.Set("Accept", "text/html")
	rec := httptest.NewRecorder()
	newTestMux().ServeHTTP(rec, req)
	csp := rec.Header().Get("Content-Security-Policy")
	directive := func(name string) string {
		for _, d := range strings.Split(csp, ";") {
			if strings.HasPrefix(strings.TrimSpace(d), name+" ") {
				return d
			}
		}
		return ""
	}
	hashOf := func(block string) string {
		sum := sha256.Sum256([]byte(block))
		return "'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
	}
	body := rec.Body.String()
	scripts := regexp.MustCompile(`(?s)<script(?:\s[^>]*)?>(.*?)</script>`).FindAllStringSubmatch(body, -1)
	if len(scripts) == 0 {
		t.Fatal("no inline scripts found in the UI")
	}
	for _, m := range scripts {
		if !strings.Contains(directive("script-src"), hashOf(m[1])) {
			t.Errorf("script-src %q lacks %s", directive("script-src"), hashOf(m[1]))
		}
	}
	for _, m := range regexp.MustCompile(`(?s)<style(?:\s[^>]*)?>(.*?)</style>`).FindAllStringSubmatch(body, -1) {
		if !strings.Contains(directive("style-src-elem"), hashOf(m[1])) {
			t.Errorf("style-src-elem %q lacks %s", directive("style-src-elem"), hashOf(m[1]))
		}
	}
	if strings.Contains(directive("script-src"), "unsafe-inline") || !strings.Contains(directive("frame-ancestors"), "'none'") || !strings.Contains(directive("connect-src"), "'self'") {
		t.Errorf("CSP = %q", csp)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
}

func TestServerServesTheUIWhateverTheAcceptHeader(t *testing.T) {
	for _, accept := range []string{"*/*", ""} {
		req := newRequest(http.MethodGet, "/", "")
		req.Header.Set("Accept", accept)
		rec := httptest.NewRecorder()
		newTestMux().ServeHTTP(rec, req)
		if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
			t.Errorf("Accept %q: Content-Type = %q, want text/html", accept, got)
		}
	}
}
