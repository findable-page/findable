package audit

import "testing"

func TestQueryFanoutCoversAddressedIntents(t *testing.T) {
	page := `<!DOCTYPE html><html><body>
	<h1>What is offline SEO scanning?</h1>
	<p>Offline SEO scanning is a local audit.</p>
	<h2>How to run a scan</h2>
	<p>Follow this step by step.</p>
	<h2>Pros and cons</h2>
	<p>The advantages are speed. The disadvantages are coverage.</p>
	<table><tr><td>findable vs alternatives</td></tr></table>
	<p>Pricing starts at zero. For example, a laptop is enough.</p>
	<h2>Common problems</h2>
	<p>One pitfall is stale HTML.</p>
	<h2>Best practices for local audits</h2>
	<p>Tips for repeatable scans.</p>
	</body></html>`
	report := QueryFanout("https://example.com", page, "offline SEO scanning")
	if report.Total != 8 || report.Covered != 8 {
		t.Fatalf("covered %d/%d, want 8/8", report.Covered, report.Total)
	}
	for _, intent := range report.Intents {
		if !intent.Covered || intent.Where == "" || intent.Query == "" {
			t.Errorf("intent %+v should be covered", intent)
		}
	}
}

func TestQueryFanoutLeavesMissingIntentsUncovered(t *testing.T) {
	page := `<!DOCTYPE html><html><body><h1>What is findable?</h1><p>Findable is a scanner.</p></body></html>`
	report := QueryFanout("file.html", page, "findable")
	got := map[string]bool{}
	for _, intent := range report.Intents {
		got[intent.ID] = intent.Covered
	}
	if !got["definition"] {
		t.Fatal("definition should be covered")
	}
	for _, id := range []string{"pricing", "examples", "problems", "practices"} {
		if got[id] {
			t.Errorf("%s should be uncovered", id)
		}
	}
}
