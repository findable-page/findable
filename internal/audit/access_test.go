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

func TestAccessGateTreatsNoneAndCrawlerSpecificDirectivesAsBlocks(t *testing.T) {
	none := `<!DOCTYPE html><html><head><meta name="robots" content="none"><title>None</title></head><body><h1>None</h1></body></html>`
	scan := Analyze("https://example.com", none, "", make(http.Header), http.StatusOK)
	assertBlocked(t, scan, "robots_none")
	if c := findCheck(scan.SEO.Checks, "robots_meta"); c == nil || c.Pass {
		t.Fatalf("robots none is noindex for the scored check, got %+v", c)
	}

	search := botByPurpose(PurposeSearch)
	directed := `<!DOCTYPE html><html><head><meta name="` + search.Name + `" content="noindex"><title>Bot</title></head><body><h1>Bot</h1></body></html>`
	scan = Analyze("https://example.com", directed, "", make(http.Header), http.StatusOK)
	assertBlocked(t, scan, "robots-"+strings.ToLower(search.Name))

	headers := make(http.Header)
	headers.Set("X-Robots-Tag", "GPTBot: noindex")
	scan = Analyze("https://example.com", `<!DOCTYPE html><html><head><title>Train</title></head><body><h1>Train</h1></body></html>`, "", headers, http.StatusOK)
	if scan.Access == nil || scan.Access.AccessBlocked {
		t.Fatalf("training-crawler noindex closed the gate: %+v", scan.Access)
	}
}

func TestAccessReasonsDescribeTheBlock(t *testing.T) {
	noindex := `<!DOCTYPE html><html><head><meta name="robots" content="noindex, nofollow"><title>Hidden</title></head><body><h1>Hidden</h1></body></html>`
	scan := Analyze("https://example.com", noindex, "", make(http.Header), http.StatusOK)
	detail := reasonDetail(t, scan, "robots_meta")
	if !strings.Contains(detail, "noindex") || strings.Contains(detail, "No robots meta tag blocking") {
		t.Fatalf("robots_meta detail = %q", detail)
	}

	nosnippet := `<!DOCTYPE html><html><head><meta name="robots" content="nosnippet"><title>Snip</title></head><body><h1>Snip</h1></body></html>`
	scan = Analyze("https://example.com", nosnippet, "", make(http.Header), http.StatusOK)
	detail = reasonDetail(t, scan, "geo_snippet_control")
	if !strings.Contains(strings.ToLower(detail), "nosnippet") || strings.Contains(detail, "No restrictive") {
		t.Fatalf("geo_snippet_control detail = %q", detail)
	}

	headers := make(http.Header)
	headers.Set("X-Robots-Tag", "noindex")
	scan = Analyze("https://example.com", `<!DOCTYPE html><html><head><title>T</title></head><body><h1>T</h1></body></html>`, "", headers, http.StatusOK)
	detail = reasonDetail(t, scan, "x_robots_tag_noindex")
	if !strings.Contains(strings.ToLower(detail), "noindex") || strings.Contains(detail, "does not noindex") {
		t.Fatalf("x_robots_tag detail = %q", detail)
	}

	search := botByPurpose(PurposeSearch)
	blocked := "User-agent: " + search.Name + "\nDisallow: /\n"
	page := `<!DOCTYPE html><html><head><title>News page with a long enough title</title></head><body><h1>News</h1></body></html>`
	scan = Analyze("https://www.bbc.com/news", page, "", make(http.Header), http.StatusOK)
	site := SiteCategory("/news", &blocked, nil)
	scan.Site = &site
	scan.RefreshAccess()
	id := "robots-" + strings.ToLower(search.Name)
	detail = reasonDetail(t, scan, id)
	if !strings.Contains(detail, "disallows") || strings.Contains(detail, "lets") {
		t.Fatalf("search crawler detail = %q", detail)
	}
}

func reasonDetail(t *testing.T, scan Scan, id string) string {
	t.Helper()
	if scan.Access == nil {
		t.Fatal("missing access gate")
	}
	for _, reason := range scan.Access.Reasons {
		if reason.ID == id {
			return reason.Detail
		}
	}
	t.Fatalf("reasons %+v missing %s", scan.Access.Reasons, id)
	return ""
}

func TestAccessGateUsesSiteRobotsForSearchCrawlersOnly(t *testing.T) {
	page := `<!DOCTYPE html><html><head><title>Open page with a long enough title here</title></head><body><h1>Open</h1></body></html>`
	scan := Analyze("https://example.com/docs", page, "", make(http.Header), http.StatusOK)
	training := botByPurpose(PurposeTraining)
	trainingOnly := "User-agent: " + training.Name + "\nDisallow: /\n"
	site := SiteCategory("/docs", &trainingOnly, nil)
	scan.Site = &site
	scan.RefreshAccess()
	if scan.Access == nil || scan.Access.AccessBlocked {
		t.Fatalf("training disallow closed the gate: %+v", scan.Access)
	}

	search := botByPurpose(PurposeSearch)
	searchBlocked := "User-agent: " + search.Name + "\nDisallow: /docs\n"
	site = SiteCategory("/docs", &searchBlocked, nil)
	scan.Site = &site
	scan.RefreshAccess()
	assertBlocked(t, scan, "robots-"+strings.ToLower(search.Name))
	trainingID := "robots-" + strings.ToLower(training.Name)
	for _, reason := range scan.Access.Reasons {
		if reason.ID == trainingID {
			t.Fatal("an allowed training crawler was reported as a block")
		}
	}
}

func botByPurpose(purpose string) Bot {
	for _, bot := range Bots {
		if bot.Purpose == purpose {
			return bot
		}
	}
	return Bot{}
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
