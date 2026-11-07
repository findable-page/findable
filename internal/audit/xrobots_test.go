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
		{"*", true},          // wildcard noindex
		{"Googlebot", true},  // inherits wildcard
		{"GPTBot", true},     // explicit
		{"ClaudeBot", false}, // own group exists, no noindex token
	}
	for _, tc := range cases {
		if got := rules.NoIndexFor(tc.agent); got != tc.want {
			t.Errorf("NoIndexFor(%s) = %v, want %v; rules %+v", tc.agent, got, tc.want, rules)
		}
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

func TestParseMetaRobotsIgnoresNonAgents(t *testing.T) {
	rules := ParseMetaRobots(`<meta name="viewport" content="noindex">`)
	if len(rules) != 0 {
		t.Errorf("viewport meta must not become a robots rule: %+v", rules)
	}
}

func TestXRobotsNoIndexOverrides(t *testing.T) {
	// A bot-specific group overrides the wildcard, mirroring robots.txt.
	rules := ParseXRobotsTag([]string{"noindex, GPTBot: index"})
	if rules.NoIndexFor("*") != true {
		t.Error("wildcard noindex lost")
	}
	if rules.NoIndexFor("GPTBot") != false {
		t.Error("GPTBot: index must override the wildcard noindex")
	}
}
