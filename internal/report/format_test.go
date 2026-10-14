package report

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nadimtuhin/findable/internal/audit"
)

const (
	examplePage = "https://example.com"
	rootPage    = "https://a.test/"
	subPage     = "https://a.test/b"
)

func TestParseFormatAcceptsKnownNamesOnly(t *testing.T) {
	for name, want := range map[string]string{"csv": ".csv", "json": ".json", "sarif": ".sarif"} {
		f, err := ParseFormat(name)
		if err != nil {
			t.Fatalf("ParseFormat(%q): %v", name, err)
		}
		if f.Ext() != want {
			t.Errorf("ParseFormat(%q).Ext() = %q, want %q", name, f.Ext(), want)
		}
	}
	if _, err := ParseFormat("xml"); err == nil {
		t.Error("ParseFormat(xml) must fail")
	}
}

func TestPageReportCSVIsTheLegacyCSV(t *testing.T) {
	s := decode(t)
	if got, want := PageReport(FormatCSV, "example.com", audit.Page{URL: examplePage, Scan: s}), CSV("example.com", s); got != want {
		t.Errorf("csv page report differs from CSV()")
	}
}

func TestPageReportJSONIsTheScanObject(t *testing.T) {
	out := PageReport(FormatJSON, "example.com", audit.Page{URL: examplePage, Scan: decode(t)})
	var got struct {
		SEO struct {
			Score  float64 `json:"score"`
			Checks []struct {
				ID string `json:"id"`
			} `json:"checks"`
		} `json:"seo"`
		CWV           json.RawMessage `json:"cwv"`
		InternalLinks json.RawMessage `json:"internalLinks"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("json report does not parse: %v\n%s", err, out)
	}
	if got.SEO.Score != 92 || len(got.SEO.Checks) != 3 || got.SEO.Checks[1].ID != "hreflang" {
		t.Errorf("json report lost scan fields: %s", out)
	}
	if got.CWV == nil || got.InternalLinks == nil {
		t.Errorf("json report must carry every API field: %s", out)
	}
}

func TestSiteReportJSONListsPagesWithURLs(t *testing.T) {
	s := decode(t)
	out := SiteReport(FormatJSON, "example.com", []audit.Page{{URL: rootPage, Scan: s}, {URL: subPage, Scan: s}})
	var got []struct {
		URL  string          `json:"url"`
		Scan json.RawMessage `json:"scan"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("site json does not parse: %v", err)
	}
	if len(got) != 2 || got[1].URL != subPage || !strings.Contains(string(got[1].Scan), `"hreflang"`) {
		t.Errorf("site json = %s", out)
	}
}

type sarifLog struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []struct {
		Tool struct {
			Driver struct {
				Name  string `json:"name"`
				Rules []struct {
					ID string `json:"id"`
				} `json:"rules"`
			} `json:"driver"`
		} `json:"tool"`
		Results []struct {
			RuleID  string `json:"ruleId"`
			Level   string `json:"level"`
			Message struct {
				Text string `json:"text"`
			} `json:"message"`
			Locations []struct {
				PhysicalLocation struct {
					ArtifactLocation struct {
						URI string `json:"uri"`
					} `json:"artifactLocation"`
				} `json:"physicalLocation"`
			} `json:"locations"`
		} `json:"results"`
	} `json:"runs"`
}

func TestSARIFReportsEachFailingCheckOnce(t *testing.T) {
	s := decode(t)
	out := SiteReport(FormatSARIF, "example.com", []audit.Page{{URL: rootPage, Scan: s}, {URL: subPage, Scan: s}})
	var log sarifLog
	if err := json.Unmarshal([]byte(out), &log); err != nil {
		t.Fatalf("sarif does not parse: %v", err)
	}
	if log.Version != "2.1.0" || !strings.Contains(log.Schema, "sarif-schema-2.1.0") || len(log.Runs) != 1 {
		t.Fatalf("sarif envelope = %s", out)
	}
	run := log.Runs[0]
	if run.Tool.Driver.Name != "findable" {
		t.Errorf("driver name = %q", run.Tool.Driver.Name)
	}
	if len(run.Tool.Driver.Rules) != 2 {
		t.Errorf("rules = %+v, want hreflang and quotes once each", run.Tool.Driver.Rules)
	}
	if len(run.Results) != 4 {
		t.Fatalf("results = %d, want 2 failing checks x 2 pages", len(run.Results))
	}
	first := run.Results[0]
	if first.RuleID != "seo/hreflang" || first.Level != "warning" || first.Message.Text != "No hreflang" {
		t.Errorf("first result = %+v", first)
	}
	if uri := first.Locations[0].PhysicalLocation.ArtifactLocation.URI; uri != rootPage {
		t.Errorf("location uri = %q", uri)
	}
}

func withSite(t *testing.T) audit.Scan {
	t.Helper()
	s := decode(t)
	s.Site = &audit.Category{Passed: 1, Total: 2, Checks: []audit.Check{
		{ID: "robots-txt", Label: "robots.txt is present", Pass: true},
		{ID: "llms-txt", Label: "/llms.txt is present", Pass: false},
	}}
	return s
}

func TestCSVAddsSiteRowsOnlyWhenSiteChecksRan(t *testing.T) {
	lines := strings.Split(CSV("example.com", withSite(t)), "\n")
	for _, want := range []string{"Site,,1,2", `robots-txt,Site,PASS,"robots.txt is present"`, `llms-txt,Site,FAIL,"/llms.txt is present"`} {
		if !contains(lines, want) {
			t.Errorf("csv is missing row %q", want)
		}
	}
	if strings.Contains(CSV("example.com", decode(t)), "Site") {
		t.Error("csv without site checks must not mention Site")
	}
}

func TestSARIFIncludesFailingSiteChecks(t *testing.T) {
	var log sarifLog
	if err := json.Unmarshal([]byte(PageReport(FormatSARIF, "example.com", audit.Page{URL: rootPage, Scan: withSite(t)})), &log); err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(log.Runs[0].Results))
	for _, r := range log.Runs[0].Results {
		ids = append(ids, r.RuleID)
	}
	if !strings.Contains(strings.Join(ids, " "), "site/llms-txt") {
		t.Errorf("sarif rule ids = %v, want site/llms-txt", ids)
	}
}
