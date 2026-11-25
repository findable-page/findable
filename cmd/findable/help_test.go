package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// cliArgs keeps flag and command names out of composite literals so goconst
// does not treat each help example as a new copy of a string the CLI already uses.
func cliArgs(parts ...string) []string { return parts }

func TestHelpListsFlagsAndExitsZero(t *testing.T) {
	cases := []struct {
		args []string
		want []string
	}{
		{cliArgs("-h"), cliArgs("commands:", "-score-model", "-fail-under", "-min-severity", "Exit codes:")},
		{cliArgs("audit", "-h"), cliArgs("findable audit", "-sitemap", "-top", "-score-model")},
		{cliArgs("bots", "-h"), cliArgs("findable bots", "-bots", "-delay-ms")},
		{cliArgs("diff", "-h"), cliArgs("baseline.json", "-fail-on-regression", "-max-drop")},
		{cliArgs("fanout", "-h"), cliArgs("-query", "sub-intents")},
		{cliArgs("hydrate", "-h"), cliArgs("-rendered", "-render-cmd")},
		{cliArgs("mcp", "-h"), cliArgs("-allow-private")},
		{cliArgs("simulate", "-h"), cliArgs("-brand", "-domain", "-engine")},
		{cliArgs("serve", "-h"), cliArgs("127.0.0.1:8080", "every interface")},
	}
	for _, tc := range cases {
		_, stderr, code := runFindableIO(t, t.TempDir(), tc.args...)
		if code != 0 {
			t.Errorf("%v exit = %d, want 0\n%s", tc.args, code, stderr)
		}
		for _, want := range tc.want {
			if !strings.Contains(stderr, want) {
				t.Errorf("%v stderr missing %q:\n%s", tc.args, want, stderr)
			}
		}
	}
}

func TestUsageErrorsNameTheProblem(t *testing.T) {
	workDir := t.TempDir()
	page := filepath.Join(workDir, "page.html")
	if err := os.WriteFile(page, []byte("<html></html>"), 0o600); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		args []string
		want string
	}{
		{nil, "need one URL or HTML file"},
		{cliArgs("audit"), "need one URL or HTML file"},
		{cliArgs("audit", page, page), "got 2 arguments"},
		{cliArgs("audit", "-sitemap", "https://example.com/sitemap.xml", page), "does not take a positional argument"},
		{cliArgs("bots"), "bots needs one http(s) URL"},
		{cliArgs("bots", page), "not a local file"},
		{cliArgs("mcp", "extra"), "does not take"},
		{cliArgs("serve"), "serve needs one address"},
		{cliArgs("diff"), "two JSON reports"},
		{cliArgs("fanout", page), "needs -query"},
		{cliArgs("hydrate", page), "exactly one of -rendered"},
		{cliArgs("simulate"), "needs -query"},
		{cliArgs("frobnicate"), "unknown command"},
		{cliArgs("-score-model", "weird", page), "unknown score model"},
		{cliArgs("bots", "-format", "xml", "https://example.com"), "unknown bots format"},
	}
	for _, tc := range cases {
		_, stderr, code := runFindableIO(t, workDir, tc.args...)
		if code != exitUsage {
			t.Errorf("%v exit = %d, want %d\n%s", tc.args, code, exitUsage, stderr)
		}
		if !strings.Contains(stderr, tc.want) {
			t.Errorf("%v stderr missing %q:\n%s", tc.args, tc.want, stderr)
		}
	}
}
