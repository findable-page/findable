package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/findable-page/findable/internal/fetch"
)

func TestClientRefusesPrivateTargetsUnlessAllowed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html><title>x</title></html>"))
	}))
	t.Cleanup(srv.Close)

	if _, err := fetch.Scan(newClient(false), srv.URL, ""); !errors.Is(err, fetch.ErrBlockedAddress) {
		t.Fatalf("default client = %v, want ErrBlockedAddress", err)
	}
	if _, err := fetch.Scan(newClient(true), srv.URL, ""); err != nil {
		t.Fatalf("-allow-private client = %v, want success", err)
	}
}

func TestGuardedCLIClientKeepsSixtySecondTimeout(t *testing.T) {
	if got := newClient(false).Timeout; got != 60*time.Second {
		t.Fatalf("guarded CLI client timeout = %v, want 60s", got)
	}
}
