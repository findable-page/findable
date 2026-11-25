package audit

import "testing"

func TestParseXRobotsTagForms(t *testing.T) {
	rules := ParseXRobotsTag([]string{
		"noindex, nofollow",
		"GPTBot: noindex, noarchive",
		"claudebot:nosnippet",
	})
	cases := []struct {
		agent string
		want  bool
	}{
		{"*", true},         // wildcard noindex
		{"Googlebot", true}, // inherits wildcard
		{"GPTBot", true},    // explicit
		{"ClaudeBot", true}, // wildcard noindex still applies
	}
	for _, tc := range cases {
		if got := rules.NoIndexFor(tc.agent); got != tc.want {
			t.Errorf("NoIndexFor(%s) = %v, want %v; rules %+v", tc.agent, got, tc.want, rules)
		}
	}
}

func TestParseXRobotsTagDirectiveBeforeNoindex(t *testing.T) {
	for _, header := range []string{"max-snippet:20, noindex", "max-image-preview:large, noindex", "unavailable_after: 2026-01-01, noindex"} {
		rules := ParseXRobotsTag([]string{header})
		if !rules.NoIndexFor("Googlebot") {
			t.Errorf("%q lost generic noindex: %+v", header, rules)
		}
	}
}

func TestParseXRobotsTagDirectiveBetweenGroups(t *testing.T) {
	rules := ParseXRobotsTag([]string{"GPTBot: index, max-snippet:20, noindex"})
	if !rules.NoIndexFor("Googlebot") {
		t.Fatalf("generic noindex lost after agent and directive: %+v", rules)
	}
}

func TestParseXRobotsTagMultipleHeaders(t *testing.T) {
	rules := ParseXRobotsTag([]string{"GPTBot: index", "max-video-preview:10, noindex"})
	if !rules.NoIndexFor("GPTBot") || !rules.NoIndexFor("Googlebot") {
		t.Fatalf("header values must combine: %+v", rules)
	}
}

func TestParseXRobotsTagUnavailableAfter(t *testing.T) {
	// "unavailable_after: date" is a real directive; the colon must not turn
	// it into an agent group that hides wildcard noindex.
	rules := ParseXRobotsTag([]string{"noindex, unavailable_after: 2026-01-01"})
	if !rules.NoIndexFor("Googlebot") {
		t.Fatalf("unavailable_after swallowed the wildcard noindex: %+v", rules)
	}
}

func TestParseMetaRobotsPerAgent(t *testing.T) {
	html := `<html><head>
	<meta name="robots" content="index, follow">
	<meta name="GPTBot" content="noindex">
	<meta name="description" content="not robots">
	</head><body>hi</body></html>`
	rules := ParseMetaRobots(html)
	if rules.NoIndexFor("*") {
		t.Error("generic robots is index; must not noindex")
	}
	if !rules.NoIndexFor("GPTBot") {
		t.Error("meta name=GPTBot noindex must apply to GPTBot")
	}
	if !rules.NoIndexFor("gptbot") {
		t.Error("agent lookup must be case-insensitive")
	}
	if rules.NoIndexFor("Googlebot") {
		t.Error("Googlebot inherits the generic robots group (index, follow), must not be noindex")
	}
}

func TestParseMetaRobotsGenericAndSpecific(t *testing.T) {
	rules := ParseMetaRobots(`<meta name="robots" content="noindex"><meta name="googlebot" content="max-snippet:20">`)
	if !rules.NoIndexFor("googlebot") {
		t.Fatalf("agent-specific directive must not cancel generic noindex: %+v", rules)
	}
}

func TestParseMetaRobotsIgnoresNonAgents(t *testing.T) {
	rules := ParseMetaRobots(`<meta name="viewport" content="noindex">`)
	if len(rules) != 0 {
		t.Errorf("viewport meta must not become a robots rule: %+v", rules)
	}
}

func TestXRobotsNoIndexCombines(t *testing.T) {
	// A bot-specific index cannot cancel generic noindex.
	rules := ParseXRobotsTag([]string{"noindex, GPTBot: index"})
	if rules.NoIndexFor("*") != true {
		t.Error("wildcard noindex lost")
	}
	if !rules.NoIndexFor("GPTBot") {
		t.Error("GPTBot: index must not override wildcard noindex")
	}
}
