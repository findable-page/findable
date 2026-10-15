package audit

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type checkCase struct {
	URL        string `json:"url"`
	Head       string `json:"head"`
	Body       string `json:"body"`
	Doc        string `json:"doc"`
	Pad        int    `json:"pad"`
	Repeat     int    `json:"repeat"`
	HeadRepeat int    `json:"headRepeat"`
}

func repeated(s string, n int) string {
	if n <= 1 {
		return s
	}
	return strings.Repeat(s, n)
}

func (c checkCase) page() (string, string) {
	url := c.URL
	if url == "" {
		url = "https://example.com/page"
	}
	if c.Doc != "" {
		return url, c.Doc
	}
	head := repeated(c.Head, c.HeadRepeat)
	if c.Pad > 0 {
		head = "<title>" + strings.Repeat("x", c.Pad) + "</title>" + head
	}
	return url, "<!DOCTYPE html><html lang='en'><head>" + head + "</head><body>" + repeated(c.Body, c.Repeat) + "</body></html>"
}

func TestSharedCheckCases(t *testing.T) {
	files, err := filepath.Glob("../../testdata/check_cases*.json")
	if err != nil || len(files) == 0 {
		t.Fatalf("no check case files: %v", err)
	}
	table := map[string]struct {
		Pass []checkCase `json:"pass"`
		Fail []checkCase `json:"fail"`
	}{}
	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, &table); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
	}
	for id, cases := range table {
		for want, group := range map[bool][]checkCase{true: cases.Pass, false: cases.Fail} {
			for i, c := range group {
				url, page := c.page()
				scan := Analyze(url, page, "", make(http.Header), 200)
				all := append(append(append([]Check{}, scan.SEO.Checks...), scan.AEO.Checks...), scan.GEO.Checks...)
				check := findCheck(all, id)
				if check == nil {
					t.Fatalf("%s: check missing", id)
				}
				if check.Pass != want {
					t.Errorf("%s case %d (want pass=%v): got pass=%v label=%q\n%s", id, i, want, check.Pass, check.Label, page)
				}
			}
		}
	}
}

func TestCheckIDsAreUnique(t *testing.T) {
	scan := Analyze("https://example.com/", "<html></html>", "kw", make(http.Header), 200)
	seen := map[string]bool{}
	for _, cat := range []Category{scan.SEO, scan.AEO, scan.GEO} {
		for _, c := range cat.Checks {
			if seen[c.ID] {
				t.Errorf("duplicate check id %s", c.ID)
			}
			seen[c.ID] = true
		}
	}
}

func TestColorContrastHostileStyleIsFast(t *testing.T) {
	const size = 1 << 20
	pages := map[string]string{
		"endless declarations": `<p style="` + strings.Repeat("color:#777;", size/11) + `">x</p>`,
		"many styled elements": strings.Repeat("<p style='color:#777;background-color:#888'>x</p>", size/50),
	}
	for name, body := range pages {
		started := time.Now()
		scan := Analyze("https://example.com/", body, "", make(http.Header), 200)
		if elapsed := time.Since(started); elapsed > time.Second {
			t.Errorf("%s: took %v", name, elapsed)
		}
		if len(scan.SEO.Checks) == 0 {
			t.Errorf("%s: no checks", name)
		}
	}
}
