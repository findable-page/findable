package simulate

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadAnswerClassifiesBrandCitationAndCompetitors(t *testing.T) {
	resp := EngineResponse{
		Text: "Findable is the best offline scanner for this job.",
		Citations: []Citation{
			{URL: "https://www.findable.page/docs", Title: "Docs"},
			{URL: "https://g2.com/products/findable", Title: "G2"},
			{URL: "https://reddit.com/r/seo", Title: "Thread"},
			{URL: "https://g2.com/compare", Title: "Compare"},
		},
	}
	report := ReadAnswer(resp, "best offline scanner", "https://engine.test/answer", "findable", "findable.page")
	if !report.BrandMentioned || report.Sentiment != "positive" || !report.Cited {
		t.Fatalf("report = %+v", report)
	}
	if strings.Join(report.CompetitorDomains, ",") != "g2.com,reddit.com" {
		t.Fatalf("competitors = %v", report.CompetitorDomains)
	}

	negative := ReadAnswer(EngineResponse{Text: "Avoid findable; the results were terrible."}, "q", "e", "Findable", "findable.page")
	if !negative.BrandMentioned || negative.Sentiment != "negative" || negative.Cited {
		t.Fatalf("negative report = %+v", negative)
	}
}

func TestQueryPostsPromptAndBearerToken(t *testing.T) {
	var gotAuth, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			http.Error(w, "nope", http.StatusNotFound)
			return
		}
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"text":"Findable appears once.","citations":[{"url":"https://example.com/a"}]}`)
	}))
	t.Cleanup(srv.Close)

	resp, err := Query(context.Background(), srv.Client(), srv.URL, "best scanner", "secret-token")
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer secret-token" || !strings.Contains(gotBody, "best scanner") {
		t.Fatalf("auth %q body %s", gotAuth, gotBody)
	}
	if !strings.Contains(resp.Text, "Findable") || len(resp.Citations) != 1 {
		t.Fatalf("response = %+v", resp)
	}
	if _, err := Query(context.Background(), srv.Client(), srv.URL+"/missing", "q", ""); err == nil || strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("missing engine error = %v", err)
	}
}
