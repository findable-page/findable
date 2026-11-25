package simulate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const engineTimeout = 30 * time.Second

// Citation is one source the engine attached to an answer.
type Citation struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

// EngineResponse is the answer payload a simulate provider returns.
type EngineResponse struct {
	Text      string     `json:"text"`
	Citations []Citation `json:"citations"`
}

// Report is the brand and competitor reading of one engine answer.
type Report struct {
	Query             string   `json:"query"`
	Engine            string   `json:"engine"`
	Brand             string   `json:"brand"`
	Domain            string   `json:"domain"`
	BrandMentioned    bool     `json:"brandMentioned"`
	Sentiment         string   `json:"sentiment,omitempty"`
	Cited             bool     `json:"cited"`
	CompetitorDomains []string `json:"competitorDomains"`
}

// ReadAnswer classifies brand mentions, the target domain, and other cited hosts.
func ReadAnswer(resp EngineResponse, query, engine, brand, domain string) Report {
	report := Report{
		Query:             query,
		Engine:            engine,
		Brand:             brand,
		Domain:            domain,
		CompetitorDomains: []string{},
	}
	if mentioned, sentiment := brandSentiment(resp.Text, brand); mentioned {
		report.BrandMentioned = true
		report.Sentiment = sentiment
	}
	own := hostKey(domain)
	seen := map[string]bool{}
	for _, citation := range resp.Citations {
		host := citationHost(citation.URL)
		if host == "" {
			continue
		}
		if own != "" && (host == own || strings.HasSuffix(host, "."+own)) {
			report.Cited = true
			continue
		}
		if !seen[host] {
			seen[host] = true
			report.CompetitorDomains = append(report.CompetitorDomains, host)
		}
	}
	sort.Strings(report.CompetitorDomains)
	return report
}

func brandSentiment(text, brand string) (bool, string) {
	lower := strings.ToLower(text)
	brand = strings.ToLower(strings.TrimSpace(brand))
	if brand == "" {
		return false, ""
	}
	at := strings.Index(lower, brand)
	if at < 0 {
		return false, ""
	}
	start := at - 80
	if start < 0 {
		start = 0
	}
	end := at + len(brand) + 80
	if end > len(lower) {
		end = len(lower)
	}
	window := lower[start:end]
	positive := countWords(window, "best", "great", "leading", "excellent", "recommended")
	negative := countWords(window, "worst", "avoid", "bad", "terrible", "poor")
	switch {
	case positive > negative:
		return true, "positive"
	case negative > positive:
		return true, "negative"
	default:
		return true, "neutral"
	}
}

func countWords(text string, words ...string) int {
	count := 0
	for _, word := range words {
		if strings.Contains(text, word) {
			count++
		}
	}
	return count
}

func citationHost(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return hostKey(parsed.Hostname())
}

func hostKey(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	host = strings.TrimPrefix(host, "www.")
	return strings.TrimSuffix(host, ".")
}

// Query posts the prompt to a simulate provider and decodes its answer.
// token, when set, is sent as a bearer token and is not included in errors.
func Query(ctx context.Context, client *http.Client, engineURL, prompt, token string) (EngineResponse, error) {
	if client == nil {
		client = &http.Client{Timeout: engineTimeout}
	}
	payload, err := json.Marshal(map[string]string{"query": prompt})
	if err != nil {
		return EngineResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, engineURL, bytes.NewReader(payload))
	if err != nil {
		return EngineResponse{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return EngineResponse{}, fmt.Errorf("query engine: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return EngineResponse{}, fmt.Errorf("read engine response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return EngineResponse{}, fmt.Errorf("engine HTTP %d", resp.StatusCode)
	}
	var parsed EngineResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return EngineResponse{}, fmt.Errorf("decode engine response: %w", err)
	}
	return parsed, nil
}
