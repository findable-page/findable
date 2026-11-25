package config

import (
	"os"
	"path/filepath"
	"testing"
)

const fixture = `{
  "baseline": ".findable/baseline.json",
  "gate": { "overall": 70, "seo": 80, "geo": 60, "noNewFailures": true, "maxSeverity": "high" },
  "ignore": [
    { "check": "aeo_conversational" },
    { "check": "img_modern_format", "url": "/legacy/*" }
  ],
  "crawl": { "limit": 200, "concurrency": 4, "include": ["/docs/*"], "exclude": ["/tag/*"] },
  "bots": ["OAI-SearchBot", "PerplexityBot"]
}`

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "findable.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadRejectsUnsupportedSections(t *testing.T) {
	for _, content := range []string{fixture, `{"gate":{"overall":70}}`, `{"bots":["GPTBot"]}`, `{"ignore":[{"check":"title_length"}]}`} {
		if _, err := Load(writeConfig(t, content)); err == nil {
			t.Fatalf("unsupported config accepted: %s", content)
		}
	}
}

func TestLoadRejectsUnknownKeys(t *testing.T) {
	_, err := Load(writeConfig(t, `{"gates": {"overall": 70}}`))
	if err == nil {
		t.Fatal("unknown key accepted")
	}
}

func TestLoadRejectsBadJSON(t *testing.T) {
	if _, err := Load(writeConfig(t, `{not json`)); err == nil {
		t.Fatal("bad json accepted")
	}
}

func TestLoadMissingFileIsError(t *testing.T) {
	if _, err := Load("/nonexistent/findable.json"); err == nil {
		t.Fatal("missing file accepted")
	}
}

func TestLoadDefaultMissingFileIsEmpty(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	c, err := LoadDefault()
	if err != nil || c.Gate.Overall != 0 {
		t.Fatalf("LoadDefault = %+v, %v", c, err)
	}
}

func TestIgnores(t *testing.T) {
	const imageCheck = "img_modern_format"
	c := Config{Ignore: []Rule{{Check: "aeo_conversational"}, {Check: imageCheck, URL: "/legacy/*"}}}
	cases := []struct {
		check, url string
		want       bool
	}{
		{"aeo_conversational", "https://x.com/any", true},
		{imageCheck, "https://x.com/legacy/page", true},
		{imageCheck, "https://x.com/docs/page", false},
		{"title_length", "https://x.com/legacy/page", false},
	}
	for _, tc := range cases {
		if got := c.Ignores(tc.check, tc.url); got != tc.want {
			t.Errorf("Ignores(%s, %s) = %v, want %v", tc.check, tc.url, got, tc.want)
		}
	}
}
