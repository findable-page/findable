package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Exit 1 is reserved for the -fail-under gate; a bad invocation is usage (2).
func TestInvalidTargetCombinationsExitUsage(t *testing.T) {
	workDir := t.TempDir()
	page := filepath.Join(workDir, "page.html")
	if err := os.WriteFile(page, []byte("<html></html>"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"-site", "-o", workDir, page},
		{"-crawl", "-o", workDir, page},
		{"-sitemap", page},
		{"-o", workDir, "http://"},
	} {
		if _, code := runFindable(t, workDir, args...); code != exitUsage {
			t.Errorf("%v exit = %d, want %d", args, code, exitUsage)
		}
	}
}

// An unknown -format is a usage error, decided before any bot is probed.
func TestBotsRejectsUnknownFormatBeforeProbing(t *testing.T) {
	stdout, code := runFindable(t, t.TempDir(), "bots", "-format", "xml", "http://127.0.0.1:1/")
	if code != exitUsage || stdout != "" {
		t.Fatalf("exit = %d, stdout %q; want %d and no output", code, stdout, exitUsage)
	}
}
