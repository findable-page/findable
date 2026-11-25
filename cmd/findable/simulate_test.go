package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSimulateCommandUsesConfiguredEngine(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"text":"Findable is a great scanner.","citations":[{"url":"https://findable.page/"},{"url":"https://www.g2.com/products/x"}]}`))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("FINDABLE_SIMULATE_TOKEN", "test-token")

	out, code := runFindable(t, t.TempDir(), "simulate", "-query", "best scanner", "-brand", "findable", "-domain", "findable.page", "-engine", srv.URL)
	if code != 0 {
		t.Fatalf("exit = %d\n%s", code, out)
	}
	if !strings.Contains(out, "mentioned yes (positive)") || !strings.Contains(out, "cited yes") || !strings.Contains(out, "g2.com") {
		t.Fatalf("stdout = %s", out)
	}
	if strings.Contains(out, "test-token") {
		t.Fatal("stdout included the engine token")
	}
}
