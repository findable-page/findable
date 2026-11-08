package audit

import (
	"strings"
	"testing"
)

const (
	googlebot   = "Googlebot"
	privatePath = "/private/x"
)

const robotsBlockingAI = `User-agent: *
Disallow: /private/
Allow: /private/open$

User-agent: GPTBot
User-agent: ClaudeBot
Disallow: /

User-agent: PerplexityBot
Disallow: /*.pdf$
`

func TestRobotsAllowsUsesTheMostSpecificGroupAndLongestRule(t *testing.T) {
	cases := []struct {
		agent, path string
		want        bool
	}{
		{googlebot, "/", true},
		{googlebot, privatePath, false},
		{googlebot, "/private/open", true},
		{googlebot, "/private/open/more", false},
		{"GPTBot", "/blog", false},
		{"claudebot", "/", false},
		{"PerplexityBot", privatePath, true},
		{"PerplexityBot", "/doc.pdf", false},
		{"Google-Extended", privatePath, false},
	}
	for _, c := range cases {
		if got := RobotsAllows(robotsBlockingAI, c.agent, c.path); got != c.want {
			t.Errorf("RobotsAllows(%s, %s) = %v, want %v", c.agent, c.path, got, c.want)
		}
	}
}

func TestRobotsMatchesTheProductTokenBeforeAVersion(t *testing.T) {
	if RobotsAllows("User-agent: GPTBot/1.0\nDisallow: /\n", "GPTBot", "/") {
		t.Error("User-agent: GPTBot/1.0 must apply to GPTBot")
	}
}

func TestRobotsAllowsEverythingWhenEmptyOrDisallowIsBlank(t *testing.T) {
	for _, robots := range []string{"", "User-agent: *\nDisallow:\n", "# just a comment\n"} {
		if !RobotsAllows(robots, "GPTBot", "/any") {
			t.Errorf("robots %q must allow everything", robots)
		}
	}
}

func siteStatuses(c Category) map[string]string {
	got := map[string]string{}
	for _, k := range c.Checks {
		got[k.ID] = Status(k)
	}
	return got
}

func TestSiteCategoryReportsRobotsAndLLMSChecks(t *testing.T) {
	robots := robotsBlockingAI
	llms := "# Example\n"
	cat := SiteCategory("/blog", &robots, &llms)
	want := map[string]string{
		"robots-txt":                StatusPass,
		"robots-page":               StatusPass,
		"robots-oai-searchbot":      StatusPass,
		"robots-perplexitybot":      StatusPass,
		"robots-claude-searchbot":   StatusPass,
		"robots-googlebot":          StatusPass,
		"robots-bingbot":            StatusPass,
		"robots-chatgpt-user":       StatusPass,
		"robots-perplexity-user":    StatusPass,
		"robots-claude-user":        StatusPass,
		"robots-gptbot":             StatusFail,
		"robots-claudebot":          StatusFail,
		"robots-google-extended":    StatusPass,
		"robots-applebot-extended":  StatusPass,
		"robots-meta-externalagent": StatusPass,
		"robots-ccbot":              StatusPass,
		"llms-txt":                  StatusPass,
	}
	got := siteStatuses(cat)
	for id, status := range want {
		if got[id] != status {
			t.Errorf("%s = %q, want %q", id, got[id], status)
		}
	}
	if len(cat.Checks) != len(want) {
		t.Errorf("site category has %d checks, want %d", len(cat.Checks), len(want))
	}
	if cat.Score != nil || cat.Passed != 15 || cat.Total != 17 {
		t.Errorf("site category score/passed/total = %v/%d/%d, want unscored 15/17", cat.Score, cat.Passed, cat.Total)
	}
}

func TestSiteCategoryExplainsSearchVsTrainingCrawlers(t *testing.T) {
	// Blocks every training bot but keeps the search bots on the * rules: the
	// page must stay citable by AI search engines.
	robots := `User-agent: GPTBot
User-agent: ClaudeBot
User-agent: Google-Extended
User-agent: Applebot-Extended
User-agent: meta-externalagent
Disallow: /
`
	llms := "# Example\n"
	cat := SiteCategory("/", &robots, &llms)
	got := siteStatuses(cat)
	for _, id := range []string{"robots-oai-searchbot", "robots-perplexitybot", "robots-claude-searchbot", "robots-googlebot"} {
		if got[id] != StatusPass {
			t.Errorf("%s = %q, want PASS: blocking training bots must not block AI search", id, got[id])
		}
	}
	for _, id := range []string{"robots-gptbot", "robots-claudebot", "robots-google-extended", "robots-applebot-extended", "robots-meta-externalagent"} {
		if got[id] != StatusFail {
			t.Errorf("%s = %q, want FAIL", id, got[id])
		}
	}
}

func TestSiteCategoryLabelsCarryCrawlerAdvice(t *testing.T) {
	labels := map[string]string{}
	for _, c := range SiteCategory("/", nil, nil).Checks {
		labels[c.ID] = c.Label
	}
	if want := "Gemini training only"; !strings.Contains(labels["robots-google-extended"], want) {
		t.Errorf("google-extended label %q must say %q", labels["robots-google-extended"], want)
	}
	if want := "does not affect AI search citations"; !strings.Contains(labels["robots-gptbot"], want) {
		t.Errorf("gptbot label %q must say %q", labels["robots-gptbot"], want)
	}
	if want := "AI search retrieval"; !strings.Contains(labels["robots-oai-searchbot"], want) {
		t.Errorf("oai-searchbot label %q must say %q", labels["robots-oai-searchbot"], want)
	}
}

func TestSiteCategoryWithoutRobotsOrLLMS(t *testing.T) {
	got := siteStatuses(SiteCategory("/", nil, nil))
	if got["robots-txt"] != StatusFail || got["llms-txt"] != StatusFail {
		t.Errorf("missing files must fail their presence checks: %v", got)
	}
	if got["robots-page"] != StatusPass || got["robots-gptbot"] != StatusPass {
		t.Errorf("no robots.txt means nothing is blocked: %v", got)
	}
}

func TestReportedAddsSiteOnlyWhenPresent(t *testing.T) {
	var s Scan
	if len(s.Reported()) != 3 {
		t.Fatalf("Reported without site = %d families, want 3", len(s.Reported()))
	}
	s.Site = &Category{}
	if r := s.Reported(); len(r) != 4 || r[3].Name != "Site" {
		t.Errorf("Reported with site = %+v", r)
	}
}

func TestRobotsRecognizesTheSearchCrawlers(t *testing.T) {
	robots := "User-agent: OAI-SearchBot\nDisallow: /\n"
	if RobotsAllows(robots, "OAI-SearchBot", "/") {
		t.Error("OAI-SearchBot must match its own group")
	}
	if !RobotsAllows(robots, "Claude-SearchBot", "/") {
		t.Error("Claude-SearchBot must not inherit OAI-SearchBot's group")
	}
}

func TestRobotsIgnoresPatternsWithMoreThanTenWildcards(t *testing.T) {
	for stars, want := range map[int]bool{10: false, 11: true} {
		robots := "User-agent: *\nDisallow: /" + strings.Repeat("*", stars) + "\n"
		if got := RobotsAllows(robots, googlebot, "/x"); got != want {
			t.Errorf("%d wildcards: RobotsAllows = %v, want %v", stars, got, want)
		}
	}
}
