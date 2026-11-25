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

func TestAccessGateShapeAndNon2xxStatuses(t *testing.T) {
	page := `<!DOCTYPE html><html><head><title>Open page with a long enough title here</title></head><body><h1>Open</h1></body></html>`
	open := Analyze("https://example.com", page, "", make(http.Header), 0)
	if open.Access == nil || open.Access.Gate != "open" || open.Access.Status != "OPEN" || open.Access.AccessBlocked || open.Access.Message != "" || len(open.Access.Reasons) != 0 {
		t.Fatalf("status 0 gate = %+v", open.Access)
	}
	for _, status := range []int{http.StatusOK, 299} {
		scan := Analyze("https://example.com", page, "", make(http.Header), status)
		if scan.Access == nil || scan.Access.AccessBlocked {
			t.Fatalf("HTTP %d closed the gate: %+v", status, scan.Access)
		}
	}
	for _, status := range []int{199, http.StatusInternalServerError, http.StatusMultipleChoices} {
		scan := Analyze("https://example.com", page, "", make(http.Header), status)
		assertBlocked(t, scan, "http_status")
	}
}

func TestAccessGateRobotsPageAndNonSearchChecks(t *testing.T) {
	page := `<!DOCTYPE html><html><head><title>Open page with a long enough title here</title></head><body><h1>Open</h1></body></html>`
	scan := Analyze("https://example.com/hidden", page, "", make(http.Header), http.StatusOK)
	allowed := "User-agent: *\nDisallow:\n"
	site := SiteCategory("/hidden", &allowed, nil)
	scan.Site = &site
	scan.RefreshAccess()
	if scan.Access == nil || scan.Access.AccessBlocked {
		t.Fatalf("missing llms.txt closed the gate: %+v", scan.Access)
	}

	blocked := "User-agent: *\nDisallow: /hidden\n"
	site = SiteCategory("/hidden", &blocked, nil)
	scan.Site = &site
	scan.RefreshAccess()
	assertBlocked(t, scan, checkRobotsPage)
	userID := "robots-" + strings.ToLower(botByPurpose(PurposeUserFetch).Name)
	for _, reason := range scan.Access.Reasons {
		if reason.ID == userID {
			t.Fatalf("user-fetch crawler closed the gate: %+v", scan.Access.Reasons)
		}
	}
}

func TestAccessGateSearchCrawlerSnippetAndNone(t *testing.T) {
	search := botByPurpose(PurposeSearch)
	id := "robots-" + strings.ToLower(search.Name)
	snippet := `<!DOCTYPE html><html><head><meta name="` + search.Name + `" content="nosnippet"><title>T</title></head><body><h1>T</h1></body></html>`
	scan := Analyze("https://example.com", snippet, "", make(http.Header), http.StatusOK)
	assertBlocked(t, scan, id)

	training := botByPurpose(PurposeTraining)
	headers := make(http.Header)
	headers.Set("X-Robots-Tag", training.Name+": nosnippet")
	scan = Analyze("https://example.com", `<!DOCTYPE html><html><head><title>T</title></head><body><h1>T</h1></body></html>`, "", headers, http.StatusOK)
	if scan.Access == nil || scan.Access.AccessBlocked {
		t.Fatalf("training nosnippet closed the gate: %+v", scan.Access)
	}

	headers = make(http.Header)
	headers.Set("X-Robots-Tag", search.Name+": none")
	scan = Analyze("https://example.com", `<!DOCTYPE html><html><head><title>T</title></head><body><h1>T</h1></body></html>`, "", headers, http.StatusOK)
	assertBlocked(t, scan, id)
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
