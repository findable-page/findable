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

	jsonOut, code := runFindable(t, t.TempDir(), "simulate", "-query", "best scanner", "-brand", "findable", "-domain", "findable.page", "-engine", srv.URL, "-format", "json")
	if code != 0 || !strings.Contains(jsonOut, `"brandMentioned":true`) || !strings.Contains(jsonOut, `"cited":true`) || strings.Contains(jsonOut, "test-token") {
		t.Fatalf("json exit = %d\n%s", code, jsonOut)
	}
}

func TestSimulateUsageAndEngineFromEnv(t *testing.T) {
	if _, code := runFindable(t, t.TempDir(), "simulate", "-query", "q"); code != exitUsage {
		t.Fatalf("missing flags exit = %d, want %d", code, exitUsage)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"text":"Findable published a note.","citations":[]}`))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("FINDABLE_SIMULATE_ENGINE", srv.URL)
	t.Setenv("FINDABLE_SIMULATE_TOKEN", "")
	out, code := runFindable(t, t.TempDir(), "simulate", "-query", "note", "-brand", "findable", "-domain", "findable.page")
	if code != 0 || !strings.Contains(out, "mentioned yes (neutral)") || !strings.Contains(out, "competitors none") {
		t.Fatalf("env engine exit = %d\n%s", code, out)
	}
	if _, code := runFindable(t, t.TempDir(), "simulate", "-query", "q", "-brand", "findable", "-domain", "findable.page", "-engine", srv.URL, "-format", "yaml"); code != exitUsage {
		t.Fatalf("bad format exit = %d, want %d", code, exitUsage)
	}
}
