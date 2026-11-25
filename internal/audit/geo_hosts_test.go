package audit

import (
	"net/http"
	"testing"
)

func TestExternalHostsIgnorePortAndTrailingDot(t *testing.T) {
	page := `<!DOCTYPE html><html><body>` +
		`<a href="https://www.cdc.gov:443/flu">CDC</a>` +
		`<a href="https://nih.gov./research">NIH</a>` +
		`<a href="https://example.com:8443/self">self</a>` +
		`</body></html>`
	sig := extractSignals("https://example.com/page", page, make(http.Header), 200)
	hosts := externalHosts(&sig)
	want := map[string]bool{"cdc.gov": true, "nih.gov": true}
	if len(hosts) != len(want) {
		t.Fatalf("externalHosts = %v, want %v", hosts, want)
	}
	for _, h := range hosts {
		if !want[h] {
			t.Errorf("unexpected host %q in %v", h, hosts)
		}
		if !isAuthoritativeHost(h) {
			t.Errorf("isAuthoritativeHost(%q) = false, want true", h)
		}
	}
}
