package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/findable-page/findable/internal/audit"
)

// Bot verdicts explain why a bot cannot see a page. Values are stable CLI
// output; do not rename.
const (
	VerdictAllowed         = "allowed"
	VerdictBlockedRobots   = "blocked-by-robots"
	VerdictBlockedServer   = "blocked-by-server"
	VerdictNoIndex         = "noindex"
	VerdictDegradedContent = "degraded-content"
	VerdictError           = "error"
	VerdictUnverifiable    = "unverifiable"
)

// BotVerdict is the live truth for one bot against one URL: what robots.txt
// says, what the server actually returns, and how the content compares to a
// browser fetch.
type BotVerdict struct {
	Bot           string `json:"bot"`
	Purpose       string `json:"purpose"`
	Consequence   string `json:"consequence,omitempty"`
	RobotsAllowed bool   `json:"robotsAllowed"`
	Status        int    `json:"status"`
	Blocked       bool   `json:"blockedByServer"`
	Challenge     string `json:"challenge,omitempty"`
	Words         int    `json:"words"`
	Parity        int    `json:"contentParityPct"`
	XNoIndex      bool   `json:"xRobotsNoIndex"`
	MetaNoIndex   bool   `json:"metaRobotsNoIndex"`
	Verdict       string `json:"verdict"`
	Error         string `json:"error,omitempty"`
}

// ProbeOptions tunes the politeness and scope of a bot probe.
type ProbeOptions struct {
	Path  string        // page path to probe, default "/"
	Bots  []string      // bot names, default all
	Delay time.Duration // pause between requests, default 250ms in the CLI
}

// challengeMarkers appear in WAF/bot-management interstitials.
var challengeMarkers = []string{
	"cf-mitigated", "cf_chl", "challenge-platform", "checking your browser",
	"verify you are human", "just a moment...", "attention required",
	"captcha", "access denied", "you have been blocked", "bot ip verification",
}

// ProbeBots requests the page as each bot's published user agent (one request
// per bot, sequential, delay apart) plus one browser-UA baseline, and turns
// the responses into verdicts. Only a live probe can tell "robots.txt allows
// it" from "the WAF blocks it anyway".
func ProbeBots(client *http.Client, pageURL string, opts ProbeOptions) ([]BotVerdict, error) { //nolint:gocyclo // validation and bot-specific verdicts stay at probe seam
	base, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", pageURL, err)
	}
	if (base.Scheme != "https" && base.Scheme != "http") || base.Host == "" { //nolint:goconst // standard URL schemes
		return nil, fmt.Errorf("need an absolute http(s) URL, got %q", pageURL)
	}
	if opts.Path == "" {
		opts.Path = "/"
	}
	if !strings.HasPrefix(opts.Path, "/") || strings.HasPrefix(opts.Path, "//") {
		return nil, fmt.Errorf("path must start with a single /: %q", opts.Path)
	}
	if opts.Delay < 0 {
		return nil, errors.New("delay must not be negative")
	}
	probeURL := base.Scheme + "://" + base.Host + opts.Path

	names := opts.Bots
	if len(names) == 0 {
		for _, b := range audit.Bots {
			names = append(names, b.Name)
		}
	}

	robotsText := ""
	if body, err := getBody(client, base.Scheme+"://"+base.Host+"/robots.txt"); err == nil {
		robotsText = string(body)
	}

	baseline, err := fetchAs(client, probeURL, userAgent)
	if err != nil {
		return nil, fmt.Errorf("baseline fetch %s: %w", probeURL, err)
	}
	if baseline.status != http.StatusOK {
		return nil, fmt.Errorf("baseline fetch %s: HTTP %d%s", probeURL, baseline.status, baseline.challenge)
	}

	verdicts := make([]BotVerdict, 0, len(names))
	for _, name := range names {
		bot, ok := audit.BotByName(strings.TrimSpace(name))
		if !ok {
			verdicts = append(verdicts, BotVerdict{Bot: name, Verdict: VerdictError, Error: "unknown bot"})
			continue
		}
		if len(verdicts) > 0 && opts.Delay > 0 {
			time.Sleep(opts.Delay)
		}
		if bot.Name == "Google-Extended" || bot.Name == "Applebot-Extended" {
			allowed := audit.RobotsAllows(robotsText, bot.Name, opts.Path)
			verdict := VerdictBlockedRobots
			if allowed {
				verdict = VerdictUnverifiable
			}
			verdicts = append(verdicts, BotVerdict{Bot: bot.Name, Purpose: bot.Purpose, Consequence: bot.Consequence, RobotsAllowed: allowed, Verdict: verdict, Challenge: "robots-only token; no distinct user agent"})
			continue
		}
		verdicts = append(verdicts, probeOne(client, probeURL, opts.Path, robotsText, bot, baseline))
	}
	return verdicts, nil
}

func probeOne(client *http.Client, probeURL, path, robotsText string, bot audit.Bot, baseline fetchResult) BotVerdict {
	v := BotVerdict{
		Bot:           bot.Name,
		Purpose:       bot.Purpose,
		Consequence:   bot.Consequence,
		RobotsAllowed: audit.RobotsAllows(robotsText, bot.Name, path),
	}
	res, err := fetchAs(client, probeURL, bot.UA)
	switch {
	case err != nil:
		v.Verdict = VerdictError
		v.Error = err.Error()
		return v
	case strings.Contains(res.challenge, "bot ip verification") && baseline.challenge == "":
		v.Verdict = VerdictUnverifiable
	case (res.challenge != "" && res.challenge != baseline.challenge) || res.status == http.StatusForbidden || res.status == http.StatusTooManyRequests || res.status >= http.StatusInternalServerError:
		v.Blocked = true
		v.Verdict = VerdictBlockedServer
	case res.status != http.StatusOK:
		v.Verdict = VerdictUnverifiable
	case !v.RobotsAllowed:
		v.Verdict = VerdictBlockedRobots
	case res.xRules.NoIndexFor(bot.Name) || res.metaRules.NoIndexFor(bot.Name):
		v.Verdict = VerdictNoIndex
	default:
		v.Verdict = VerdictAllowed
	}

	v.Status = res.status
	v.Words = res.words
	if baseline.words > 0 {
		v.Parity = res.words * 100 / baseline.words
	}
	v.XNoIndex = res.xRules.NoIndexFor(bot.Name)
	v.MetaNoIndex = res.metaRules.NoIndexFor(bot.Name)
	v.Challenge = res.challenge
	return classifyParity(v)
}

func classifyParity(v BotVerdict) BotVerdict {
	if v.Verdict == VerdictAllowed && v.Parity > 0 && v.Parity < 50 {
		v.Verdict = VerdictDegradedContent
	}
	return v
}

type fetchResult struct {
	status    int
	words     int
	challenge string
	xRules    audit.XRobotsRules
	metaRules audit.XRobotsRules
}

func fetchAs(client *http.Client, target, ua string) (fetchResult, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, http.NoBody)
	if err != nil {
		return fetchResult{}, err
	}
	req.Header.Set("User-Agent", ua)
	resp, err := client.Do(req) //nolint:gosec // probing the user-named URL is the feature; see the SSRF policy in SECURITY.md
	if err != nil {
		return fetchResult{}, err
	}
	defer closeBody(resp)

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxPageBytes))
	if err != nil {
		return fetchResult{}, fmt.Errorf("read body: %w", err)
	}
	xRules := audit.ParseXRobotsTag(resp.Header.Values("X-Robots-Tag"))
	metaRules := audit.ParseMetaRobots(string(body))
	return fetchResult{
		status:    resp.StatusCode,
		words:     countWords(body),
		challenge: detectChallenge(resp, body),
		xRules:    xRules,
		metaRules: metaRules,
	}, nil
}

func countWords(body []byte) int {
	return len(strings.Fields(audit.VisibleText(string(body))))
}

// detectChallenge surfaces the WAF/bot-management evidence for a block.
func detectChallenge(resp *http.Response, body []byte) string {
	lower := strings.ToLower(string(body))
	for _, marker := range challengeMarkers {
		if (resp.StatusCode != http.StatusOK || len(body) < 2048) && strings.Contains(lower, marker) {
			return fmt.Sprintf("status %d, body contains %q", resp.StatusCode, marker)
		}
	}
	if v := resp.Header.Get("Cf-Mitigated"); v != "" {
		return "cf-mitigated: " + v
	}
	if resp.StatusCode == http.StatusForbidden && strings.EqualFold(resp.Header.Get("Server"), "cloudflare") {
		return "403 from cloudflare"
	}
	return ""
}
