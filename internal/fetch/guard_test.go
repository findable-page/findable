package fetch

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
)

func TestPublicAddrRejectsLoopbackPrivateLinkLocalAndMetadata(t *testing.T) {
	for addr, want := range map[string]bool{
		"8.8.8.8":              true,
		"2606:4700::1111":      true,
		"127.0.0.1":            false,
		"10.1.2.3":             false,
		"172.16.0.1":           false,
		"192.168.68.60":        false,
		"169.254.169.254":      false,
		"100.104.51.75":        false,
		"0.0.0.0":              false,
		"::1":                  false,
		"fd00::1":              false,
		"fe80::1":              false,
		"::ffff:127.0.0.1":     false,
		"2002:7f00:1::1":       false,
		"2001::1":              false,
		"2001:0:4136:e378::1":  false,
		"2001:4860:4860::8888": true,
		"64:ff9b::7f00:1":      false,
		"0.1.2.3":              false,
		"198.18.0.1":           false,
		"198.19.255.1":         false,
	} {
		if got := PublicAddr(netip.MustParseAddr(addr)); got != want {
			t.Errorf("PublicAddr(%s) = %v, want %v", addr, got, want)
		}
	}
}

func TestGuardedClientRefusesLoopbackTargets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html></html>"))
	}))
	t.Cleanup(srv.Close)
	_, err := Scan(NewGuardedClient(), srv.URL, "", false)
	if !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("Scan via guarded client = %v, want ErrBlockedAddress", err)
	}
}

func TestScanRefusesBodiesOverTheCap(t *testing.T) {
	big := strings.Repeat("a", MaxPageBytes+1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(big))
	}))
	t.Cleanup(srv.Close)
	if _, err := Scan(srv.Client(), srv.URL, "", false); !errors.Is(err, ErrPageTooLarge) {
		t.Fatalf("Scan of oversized page = %v, want ErrPageTooLarge", err)
	}
}

func TestGuardedClientRefusesTrailingDotLocalhost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html></html>"))
	}))
	t.Cleanup(srv.Close)
	target := strings.Replace(srv.URL, "127.0.0.1", "localhost.", 1)
	if _, err := Scan(NewGuardedClient(), target, "", false); !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("Scan(%s) via guarded client = %v, want ErrBlockedAddress", target, err)
	}
}

func TestGuardedClientRefusesNonWebPorts(t *testing.T) {
	if _, err := Scan(NewGuardedClient(), "http://8.8.8.8:8080/", "", false); !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("Scan of a public IP on port 8080 = %v, want ErrBlockedAddress", err)
	}
}
