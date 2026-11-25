package simulate

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
	if !report.BrandMentioned || report.Sentiment != sentimentPositive || !report.Cited {
		t.Fatalf("report = %+v", report)
	}
	if strings.Join(report.CompetitorDomains, ",") != "g2.com,reddit.com" {
		t.Fatalf("competitors = %v", report.CompetitorDomains)
	}

	negative := ReadAnswer(EngineResponse{Text: "Avoid findable; the results were terrible."}, "q", "e", "Findable", "findable.page")
	if !negative.BrandMentioned || negative.Sentiment != sentimentNegative || negative.Cited {
		t.Fatalf("negative report = %+v", negative)
	}
}

func TestReadAnswerIgnoresEmbeddedTermsAndNormalizesHosts(t *testing.T) {
	embedded := ReadAnswer(EngineResponse{
		Text: "An unfindable badge is not a brand mention.",
		Citations: []Citation{
			{URL: "https://notfindable.page/about"},
			{URL: "findable.page/docs"},
			{URL: "https://docs.findable.page/guide"},
		},
	}, "q", "e", "findable", "https://www.findable.page/path")
	if embedded.BrandMentioned || embedded.Sentiment != "" || !embedded.Cited {
		t.Fatalf("embedded report = %+v", embedded)
	}
	if strings.Join(embedded.CompetitorDomains, ",") != "notfindable.page" {
		t.Fatalf("competitors = %v", embedded.CompetitorDomains)
	}

	neutral := ReadAnswer(EngineResponse{Text: "Findable published a changelog."}, "q", "e", "Findable", "findable.page:443")
	if !neutral.BrandMentioned || neutral.Sentiment != sentimentNeutral {
		t.Fatalf("neutral report = %+v", neutral)
	}
}

func TestReadAnswerSentimentTieIgnoresDistantWords(t *testing.T) {
	tied := ReadAnswer(EngineResponse{Text: "Findable is the best and the worst scanner."}, "q", "e", "findable", "findable.page")
	if !tied.BrandMentioned || tied.Sentiment != sentimentNeutral {
		t.Fatalf("tie = %+v", tied)
	}
	far := "Findable. " + strings.Repeat("pad ", 80) + "terrible results"
	distant := ReadAnswer(EngineResponse{Text: far}, "q", "e", "Findable", "findable.page")
	if !distant.BrandMentioned || distant.Sentiment != sentimentNeutral {
		t.Fatalf("distant = %+v", distant)
	}
	if ReadAnswer(EngineResponse{Text: "Findable is great"}, "q", "e", "  ", "findable.page").BrandMentioned {
		t.Fatal("blank brand matched")
	}
}

func TestReadAnswerCitationHosts(t *testing.T) {
	report := ReadAnswer(EngineResponse{Citations: []Citation{
		{URL: "https://findable.page.evil.com/a"},
		{URL: "https://www.g2.com:443/products"},
		{URL: ""},
		{URL: "javascript:alert(1)"},
		{URL: "https://docs.findable.page/guide"},
	}}, "q", "e", "findable", "findable.page/docs")
	if !report.Cited {
		t.Fatal("own subdomain should count as cited")
	}
	if got := strings.Join(report.CompetitorDomains, ","); got != "findable.page.evil.com,g2.com" {
		t.Fatalf("competitors = %q", got)
	}
}

func TestRedactSecretStripsTokenAndUserinfo(t *testing.T) {
	msg := redactSecret("tok=abc pass=xyz", "abc", "http://user:xyz@engine.test/q")
	if strings.Contains(msg, "abc") || strings.Contains(msg, "xyz") {
		t.Fatalf("secret left in %q", msg)
	}
}

func TestQueryErrorOmitsToken(t *testing.T) {
	const token = "secret-token"
	_, err := Query(context.Background(), &http.Client{Timeout: time.Second}, "http://127.0.0.1:1/?access_token="+token, "q", token)
	if err == nil || strings.Contains(err.Error(), token) {
		t.Fatal("engine error included the token or did not fail")
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
