package fetch

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/findable-page/findable/internal/audit"
)

const googlebotName = "Googlebot"

var browserPage = `<html><head><title>Real page</title></head><body>
<main>` + strings.Repeat("lorem ipsum dolor ", 100) + `</main></body></html>`

var challengePage = `<html><body>Just a moment... Checking your browser before you access</body></html>`

// botProbeServer fakes a WAF: GPTBot gets 403 + cf-mitigated, PerplexityBot
// gets a tiny challenge page with status 200, ClaudeBot gets a per-bot
// noindex X-Robots-Tag, everyone else gets the real page.
func botProbeServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("User-agent: GPTBot\nDisallow: /\n\nUser-agent: *\nAllow: /\n"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.UserAgent(), "OAI-SearchBot"):
			_, _ = w.Write([]byte(`<html><head><meta name="robots" content="noindex"></head><body>` + strings.Repeat("word ", 60) + `</body></html>`))
		case strings.Contains(r.UserAgent(), "GPTBot"):
			w.Header().Set("Cf-Mitigated", "challenge")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(challengePage))
		case strings.Contains(r.UserAgent(), "PerplexityBot"):
			_, _ = w.Write([]byte(challengePage))
		case strings.Contains(r.UserAgent(), "ClaudeBot"):
			w.Header().Set("X-Robots-Tag", "ClaudeBot: noindex")
			_, _ = w.Write([]byte(browserPage))
		default:
			_, _ = w.Write([]byte(browserPage))
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func probe(t *testing.T, srv *httptest.Server, bots []string) []BotVerdict {
	t.Helper()
	client := &http.Client{Timeout: 5 * time.Second} //nolint:noctx // test fixture
	verdicts, err := ProbeBots(client, srv.URL+"/page", ProbeOptions{Bots: bots, Delay: 0})
	if err != nil {
		t.Fatal(err)
	}
	return verdicts
}

func byBot(verdicts []BotVerdict, name string) BotVerdict {
	for _, v := range verdicts {
		if v.Bot == name {
			return v
		}
	}
	return BotVerdict{}
}

func TestProbeBotsVerdicts(t *testing.T) {
	srv := botProbeServer(t)
	verdicts := probe(t, srv, []string{"OAI-SearchBot", "GPTBot", "PerplexityBot", "ClaudeBot", googlebotName})

	cases := []struct {
		bot   string
		want  string
		check func(BotVerdict) bool
		why   string
	}{
		{"OAI-SearchBot", VerdictNoIndex, func(v BotVerdict) bool { return v.MetaNoIndex }, "meta robots noindex"},
		{"GPTBot", VerdictBlockedServer, func(v BotVerdict) bool { return v.Challenge != "" && !v.RobotsAllowed }, "403 + cf-mitigated; robots also disallows"},
		{"PerplexityBot", VerdictBlockedServer, func(v BotVerdict) bool { return v.Parity < 50 && v.Challenge != "" }, "challenge body has far fewer words than baseline"},
		{"ClaudeBot", VerdictNoIndex, func(v BotVerdict) bool { return v.XNoIndex }, "X-Robots-Tag ClaudeBot: noindex"},
		{googlebotName, VerdictAllowed, func(v BotVerdict) bool { return v.Status == http.StatusOK }, "normal fetch"},
	}
	for _, tc := range cases {
		v := byBot(verdicts, tc.bot)
		if v.Verdict != tc.want || !tc.check(v) {
			t.Errorf("%s = %q (challenge %q, parity %d, xNoIndex %v, metaNoIndex %v, robots %v), want %q because %s",
				tc.bot, v.Verdict, v.Challenge, v.Parity, v.XNoIndex, v.MetaNoIndex, v.RobotsAllowed, tc.want, tc.why)
		}
	}
}

func TestProbeBotsSequentialPoliteness(t *testing.T) {
	srv := botProbeServer(t)
	start := time.Now()
	verdicts := probe(t, srv, []string{"Googlebot", "Bingbot"})
	if elapsed := time.Since(start); elapsed < 2*50*time.Millisecond {
		t.Errorf("probe of 2 bots took %v; delay between requests not honored", elapsed)
	}
	if len(verdicts) != 2 {
		t.Fatalf("verdicts = %d", len(verdicts))
	}
}

func TestProbeBotsUnknownBot(t *testing.T) {
	srv := botProbeServer(t)
	verdicts := probe(t, srv, []string{"NotABot"})
	v := verdicts[0]
	if v.Verdict != VerdictError || v.Error == "" {
		t.Fatalf("unknown bot verdict = %+v", v)
	}
}

func TestProbeBotsPathUsesOrigin(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		_, _ = w.Write([]byte(browserPage))
	}))
	defer srv.Close()
	client := &http.Client{Timeout: time.Second} //nolint:noctx // test fixture
	_, err := ProbeBots(client, srv.URL+"/article", ProbeOptions{Path: "/pricing", Bots: []string{googlebotName}, Delay: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if path != "/robots.txt" && path != "/pricing" {
			t.Fatalf("requested %q instead of /pricing; all paths %v", path, paths)
		}
	}
}

func TestProbeBotsNeedsAbsoluteURL(t *testing.T) {
	client := &http.Client{} //nolint:noctx // test fixture
	if _, err := ProbeBots(client, "example.com", ProbeOptions{}); err == nil {
		t.Fatal("relative URL accepted")
	}
}

func TestBotTableDrivesSiteChecks(t *testing.T) {
	// every bot in the table must produce a robots-<bot> site check ID
	robots := "User-agent: *\nAllow: /\n"
	cat := audit.SiteCategory("/", &robots, nil)
	seen := map[string]bool{}
	for _, k := range cat.Checks {
		seen[k.ID] = true
	}
	for _, b := range audit.Bots {
		id := "robots-" + strings.ToLower(b.Name)
		if !seen[id] {
			t.Errorf("bot %s missing from site checks (want %s)", b.Name, id)
		}
	}
}
