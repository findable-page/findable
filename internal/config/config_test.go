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

func TestLoadParsesAllSections(t *testing.T) {
	c, err := Load(writeConfig(t, fixture))
	if err != nil {
		t.Fatal(err)
	}
	if c.Baseline != ".findable/baseline.json" {
		t.Errorf("baseline = %q", c.Baseline)
	}
	if c.Gate.Overall != 70 || c.Gate.SEO != 80 || c.Gate.GEO != 60 || !c.Gate.NoNewFailures || c.Gate.MaxSeverity != "high" {
		t.Errorf("gate = %+v", c.Gate)
	}
	if len(c.Ignore) != 2 || c.Ignore[1].URL != "/legacy/*" {
		t.Errorf("ignore = %+v", c.Ignore)
	}
	if c.Crawl.Limit != 200 || c.Crawl.Concurrency != 4 || len(c.Crawl.Include) != 1 || len(c.Crawl.Exclude) != 1 {
		t.Errorf("crawl = %+v", c.Crawl)
	}
	if len(c.Bots) != 2 {
		t.Errorf("bots = %+v", c.Bots)
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
	wd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Skip("cannot chdir")
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	c, err := LoadDefault()
	if err != nil || c.Gate.Overall != 0 {
		t.Fatalf("LoadDefault = %+v, %v", c, err)
	}
}

func TestIgnores(t *testing.T) {
	c, err := Load(writeConfig(t, fixture))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		check, url string
		want       bool
	}{
		{"aeo_conversational", "https://x.com/any", true},
		{"img_modern_format", "https://x.com/legacy/page", true},
		{"img_modern_format", "https://x.com/docs/page", false},
		{"title_length", "https://x.com/legacy/page", false},
	}
	for _, tc := range cases {
		if got := c.Ignores(tc.check, tc.url); got != tc.want {
			t.Errorf("Ignores(%s, %s) = %v, want %v", tc.check, tc.url, got, tc.want)
		}
	}
}
