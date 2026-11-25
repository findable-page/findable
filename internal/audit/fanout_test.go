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

func TestQueryFanoutDoesNotTreatEveryTableAsComparison(t *testing.T) {
	page := `<!DOCTYPE html><html><body><table><tr><td>Monday</td><td>Tuesday</td></tr></table></body></html>`
	report := QueryFanout("https://example.com", page, "findable")
	for _, intent := range report.Intents {
		if intent.ID == "comparison" && intent.Covered {
			t.Fatal("a calendar table covered comparison")
		}
	}
}

func TestQueryFanoutNamesABlankTopic(t *testing.T) {
	report := QueryFanout("https://example.com", `<!DOCTYPE html><html><body></body></html>`, "  ")
	if report.Topic != "" || report.Total != 8 || report.Covered != 0 {
		t.Fatalf("blank topic report = %+v", report)
	}
	if report.Intents[0].Query != "what is this topic" || report.Intents[0].Where != "" {
		t.Fatalf("definition intent = %+v", report.Intents[0])
	}
}

func TestQueryFanoutReportsTheFirstMatchingSpot(t *testing.T) {
	page := `<!DOCTYPE html><html><body>
	<h2>Pricing</h2>
	<p>The price is also in the paragraph. e.g. a free tier.</p>
	<table><tr><td>For example, a row.</td></tr></table>
	</body></html>`
	report := QueryFanout("https://example.com", page, "findable")
	where := map[string]string{}
	covered := map[string]bool{}
	for _, intent := range report.Intents {
		where[intent.ID] = intent.Where
		covered[intent.ID] = intent.Covered
	}
	if where["pricing"] != "h2" {
		t.Fatalf("pricing where = %q", where["pricing"])
	}
	if !covered["examples"] || where["examples"] != "p:1" {
		t.Fatalf("examples intent where = %q covered = %v", where["examples"], covered["examples"])
	}
}

func TestQueryFanoutDoesNotMatchIntentWordsInsideOtherWords(t *testing.T) {
	embedded := `<!DOCTYPE html><html><body>
	<h2>Priceless advice</h2>
	<p>Alternatively, consider the accost of a stranger.</p>
	</body></html>`
	report := QueryFanout("https://example.com", embedded, "findable")
	for _, intent := range report.Intents {
		if intent.Covered && (intent.ID == "pricing" || intent.ID == "comparison") {
			t.Fatalf("%s covered by an embedded word at %s", intent.ID, intent.Where)
		}
	}

	plural := `<!DOCTYPE html><html><body><h2>Alternatives</h2></body></html>`
	report = QueryFanout("https://example.com", plural, "findable")
	for _, intent := range report.Intents {
		if intent.ID == "comparison" && !intent.Covered {
			t.Fatal("alternatives should cover comparison")
		}
	}
}
