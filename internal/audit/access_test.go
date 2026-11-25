package audit

import (
	"net/http"
	"strings"
	"testing"
)

func TestAccessGateStaysOpenForAnIndexablePage(t *testing.T) {
	page := `<!DOCTYPE html><html><head><title>Open page with a long enough title here</title></head><body><h1>Open</h1></body></html>`
	scan := Analyze("https://example.com", page, "", make(http.Header), http.StatusOK)
	if scan.Access == nil || scan.Access.AccessBlocked || scan.Access.Status != "OPEN" {
		t.Fatalf("indexable page access = %+v", scan.Access)
	}
}

func TestAccessGateBlocksNoindexNosnippetAndHTTPErrors(t *testing.T) {
	noindex := `<!DOCTYPE html><html><head><meta name="robots" content="noindex"></head><body><h1>Hidden</h1></body></html>`
	scan := Analyze("https://example.com", noindex, "", make(http.Header), http.StatusOK)
	assertBlocked(t, scan, "robots_meta")

	nosnippet := `<!DOCTYPE html><html><head><meta name="robots" content="nosnippet"></head><body><h1>No snippet</h1></body></html>`
	scan = Analyze("https://example.com", nosnippet, "", make(http.Header), http.StatusOK)
	assertBlocked(t, scan, "geo_snippet_control")

	headers := make(http.Header)
	headers.Set("X-Robots-Tag", "noindex")
	scan = Analyze("https://example.com", `<!DOCTYPE html><html><head><title>T</title></head></html>`, "", headers, http.StatusOK)
	assertBlocked(t, scan, "x_robots_tag_noindex")

	scan = Analyze("https://example.com", `<!DOCTYPE html><html></html>`, "", make(http.Header), http.StatusNotFound)
	assertBlocked(t, scan, "http_status")
	if scan.Access.Message == "" || scan.Access.Status != "BLOCKED_FROM_INDEXING" {
		t.Fatalf("blocked gate = %+v", scan.Access)
	}
}

func TestAccessGateIncludesBlockedSearchCrawlers(t *testing.T) {
	scan := Analyze("https://example.com", `<!DOCTYPE html><html><head><title>Open</title></head></html>`, "", make(http.Header), http.StatusOK)
	trainingID := firstBotCheckID(PurposeTraining)
	searchID := firstBotCheckID(PurposeSearch)
	scan.Site = &Category{Checks: []Check{
		{ID: trainingID, Label: "training bot blocked", Pass: false},
		{ID: searchID, Label: "search bot disallowed", Pass: false},
	}}
	scan.RefreshAccess()
	assertBlocked(t, scan, searchID)
	for _, reason := range scan.Access.Reasons {
		if reason.ID == trainingID {
			t.Fatal("training crawler blocks must not close the access gate")
		}
	}
}

func firstBotCheckID(purpose string) string {
	for _, bot := range Bots {
		if bot.Purpose == purpose {
			return "robots-" + strings.ToLower(bot.Name)
		}
	}
	return ""
}

func assertBlocked(t *testing.T, scan Scan, reasonID string) {
	t.Helper()
	if scan.Access == nil || !scan.Access.AccessBlocked {
		t.Fatalf("want blocked gate, got %+v", scan.Access)
	}
	for _, reason := range scan.Access.Reasons {
		if reason.ID == reasonID {
			return
		}
	}
	t.Fatalf("reasons %+v missing %s", scan.Access.Reasons, reasonID)
}
