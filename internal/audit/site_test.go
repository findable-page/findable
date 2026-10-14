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
		"robots-txt":             StatusPass,
		"robots-page":            StatusPass,
		"robots-gptbot":          StatusFail,
		"robots-claudebot":       StatusFail,
		"robots-perplexitybot":   StatusPass,
		"robots-google-extended": StatusPass,
		"llms-txt":               StatusPass,
	}
	got := siteStatuses(cat)
	for id, status := range want {
		if got[id] != status {
			t.Errorf("%s = %q, want %q", id, got[id], status)
		}
	}
	if cat.Score != nil || cat.Passed != 5 || cat.Total != 7 {
		t.Errorf("site category score/passed/total = %v/%d/%d, want unscored 5/7", cat.Score, cat.Passed, cat.Total)
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

func TestRobotsIgnoresPatternsWithMoreThanTenWildcards(t *testing.T) {
	for stars, want := range map[int]bool{10: false, 11: true} {
		robots := "User-agent: *\nDisallow: /" + strings.Repeat("*", stars) + "\n"
		if got := RobotsAllows(robots, googlebot, "/x"); got != want {
			t.Errorf("%d wildcards: RobotsAllows = %v, want %v", stars, got, want)
		}
	}
}
