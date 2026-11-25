package audit

import (
	"net/http"
	"testing"
)

func TestClassifyLinkTreatsHostCaseAndDefaultPortAsSameSite(t *testing.T) {
	page := `<!DOCTYPE html><html><body>` +
		`<a href="https://EXAMPLE.com/about">about</a>` +
		`<a href="https://example.com:443/pricing">pricing</a>` +
		`<a href="https://example.com./blog">blog</a>` +
		`<a href="https://example.com:8443/admin">other port</a>` +
		`<a href="https://other.org/">other</a>` +
		`</body></html>`
	sig := extractSignals("https://example.com/page", page, make(http.Header), 200)
	if len(sig.InternalLinks) != 3 {
		t.Errorf("InternalLinks = %v, want the 3 same-site links", sig.InternalLinks)
	}
	if len(sig.ExternalLinks) != 2 {
		t.Errorf("ExternalLinks = %v, want :8443 and other.org", sig.ExternalLinks)
	}
}
