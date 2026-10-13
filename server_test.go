package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestMux() *http.ServeMux {
	client := &http.Client{}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if strings.Contains(r.Header.Get("Accept"), "text/html") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(localWebUIHTML))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","name":"seoscore-cli-server"}`))
	})
	mux.HandleFunc("/api/scan", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			URL       string `json:"url"`
			Keyphrase string `json:"keyphrase"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.URL) == "" {
			http.Error(w, "bad request: missing url", http.StatusBadRequest)
			return
		}
		res, err := scanTarget(client, req.URL, req.Keyphrase, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})
	mux.HandleFunc("/api/scan-count", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":2847}`))
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(`{"status":"ok"}`))
	})
	return mux
}

func TestServerHealthEndpoint(t *testing.T) {
	mux := newTestMux()
	req := httptest.NewRequest("GET", "/health", nil)
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
	req := httptest.NewRequest("GET", "/api/scan-count", nil)
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
	req := httptest.NewRequest("GET", "/", nil)
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
	req := httptest.NewRequest("OPTIONS", "/api/scan", nil)
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
	req := httptest.NewRequest("POST", "/api/scan", bytes.NewBufferString("{}"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty json, got %d", rec.Code)
	}

	// 2. Whitespace URL
	req = httptest.NewRequest("POST", "/api/scan", bytes.NewBufferString(`{"url":"   "}`))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for whitespace url, got %d", rec.Code)
	}

	// 3. Invalid JSON
	req = httptest.NewRequest("POST", "/api/scan", bytes.NewBufferString(`invalid json`))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid json, got %d", rec.Code)
	}
}

func TestServerScanLocalFile(t *testing.T) {
	mux := newTestMux()
	reqBody := `{"url":"testdata/site/01-thin-poor.html"}`
	req := httptest.NewRequest("POST", "/api/scan", bytes.NewBufferString(reqBody))
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
