package simulate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	engineTimeout       = 30 * time.Second
	sentimentPositive   = "positive"
	sentimentNegative   = "negative"
	sentimentNeutral    = "neutral"
	sentimentWindowRune = 80
)

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
	at := termIndex(lower, brand)
	if at < 0 {
		return false, ""
	}
	window := runeWindow(lower, at, len(brand), sentimentWindowRune)
	positive := countTerms(window, "best", "great", "leading", "excellent", "recommended")
	negative := countTerms(window, "worst", "avoid", "bad", "terrible", "poor")
	switch {
	case positive > negative:
		return true, sentimentPositive
	case negative > positive:
		return true, sentimentNegative
	default:
		return true, sentimentNeutral
	}
}

func countTerms(text string, words ...string) int {
	count := 0
	for _, word := range words {
		if termIndex(text, word) >= 0 {
			count++
		}
	}
	return count
}

// termIndex reports the byte offset of term when it is not glued to another
// letter or digit. "findable" does not match "unfindable", and "bad" does not
// match "badge".
func termIndex(text, term string) int {
	if term == "" {
		return -1
	}
	for start := 0; start < len(text); {
		rel := strings.Index(text[start:], term)
		if rel < 0 {
			return -1
		}
		at := start + rel
		if termBounded(text, at, at+len(term)) {
			return at
		}
		start = at + 1
	}
	return -1
}

func termBounded(text string, start, end int) bool {
	if start > 0 {
		r, _ := utf8.DecodeLastRuneInString(text[:start])
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}
	if end < len(text) {
		r, _ := utf8.DecodeRuneInString(text[end:])
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func runeWindow(text string, at, termLen, pad int) string {
	start := at
	for i := 0; i < pad && start > 0; i++ {
		_, size := utf8.DecodeLastRuneInString(text[:start])
		if size <= 0 {
			break
		}
		start -= size
	}
	end := at + termLen
	for i := 0; i < pad && end < len(text); i++ {
		_, size := utf8.DecodeRuneInString(text[end:])
		if size <= 0 {
			break
		}
		end += size
	}
	return text[start:end]
}

func citationHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	candidate := raw
	if !strings.Contains(candidate, "://") && !strings.HasPrefix(candidate, "//") {
		candidate = "https://" + candidate
	}
	parsed, err := url.Parse(candidate)
	if err != nil || parsed.Hostname() == "" {
		return ""
	}
	return hostKey(parsed.Hostname())
}

func hostKey(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.Contains(raw, "://") || strings.HasPrefix(raw, "//") {
		if parsed, err := url.Parse(raw); err == nil && parsed.Hostname() != "" {
			raw = parsed.Hostname()
		}
	}
	if slash := strings.IndexByte(raw, '/'); slash >= 0 {
		raw = raw[:slash]
	}
	raw = strings.TrimSuffix(raw, ".")
	if host, _, err := net.SplitHostPort(raw); err == nil {
		raw = host
	}
	return strings.TrimPrefix(strings.ToLower(raw), "www.")
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
		return EngineResponse{}, fmt.Errorf("create request: %s", redactSecret(err.Error(), token, engineURL))
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return EngineResponse{}, fmt.Errorf("query engine: %s", redactSecret(err.Error(), token, engineURL))
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

// redactSecret removes the bearer token and any userinfo password from a
// transport error. Engine URLs sometimes carry the credential in the query.
func redactSecret(msg, token, engineURL string) string {
	if token != "" {
		msg = strings.ReplaceAll(msg, token, "redacted")
	}
	parsed, err := url.Parse(engineURL)
	if err != nil || parsed.User == nil {
		return msg
	}
	if password, ok := parsed.User.Password(); ok && password != "" {
		msg = strings.ReplaceAll(msg, password, "redacted")
	}
	return msg
}
