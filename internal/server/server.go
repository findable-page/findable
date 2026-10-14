package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/nadimtuhin/seoscore-cli/internal/audit"
	"github.com/nadimtuhin/seoscore-cli/internal/fetch"
	"github.com/nadimtuhin/seoscore-cli/internal/report"
)

const (
	readHeaderTimeout = 10 * time.Second
	rootStatusJSON    = `{"status":"ok","name":"seoscore-cli-server"}`
	scanCountJSON     = `{"count":2847}`
	healthJSON        = `{"status":"ok"}`
	jsonContentType   = "application/json"
	scansPerWindow    = 10
	scanWindow        = time.Minute
)

type rateLimiter struct {
	mu          sync.Mutex
	limit       int
	window      time.Duration
	now         func() time.Time
	windowStart time.Time
	counts      map[string]int
}

func newRateLimiter(limit int, window time.Duration, now func() time.Time) *rateLimiter {
	return &rateLimiter{limit: limit, window: window, now: now, windowStart: now(), counts: map[string]int{}}
}

func (l *rateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if t := l.now(); t.Sub(l.windowStart) >= l.window {
		l.windowStart = t
		clear(l.counts)
	}
	l.counts[key]++
	return l.counts[key] <= l.limit
}

func clientKey(r *http.Request) string {
	ap, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	addr := ap.Addr().Unmap().WithZone("")
	if addr.Is6() {
		return netip.PrefixFrom(addr, 64).Masked().String()
	}
	return addr.String()
}

type scanRequest struct {
	URL       string `json:"url"`
	Keyphrase string `json:"keyphrase"`
	Format    string `json:"format"`
	Site      bool   `json:"site"`
}

var formatContentTypes = map[report.Format]string{
	report.FormatCSV:   "text/csv; charset=utf-8",
	report.FormatSARIF: "application/sarif+json",
}

func NewMux(client *http.Client, remote bool) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/api/scan", scanHandler(client, remote, newRateLimiter(scansPerWindow, scanWindow, time.Now)))
	mux.HandleFunc("/api/scan-count", handleScanCount)
	mux.HandleFunc("/health", handleHealth)
	return mux
}

func NewServer(addr string, remote bool) *http.Server {
	return &http.Server{Addr: addr, Handler: NewMux(fetch.NewGuardedClient(), remote), ReadHeaderTimeout: readHeaderTimeout}
}

func allowAnyOrigin(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func writeBody(w http.ResponseWriter, body string) {
	_, _ = w.Write([]byte(body))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	allowAnyOrigin(w)
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		writeBody(w, localWebUIHTML)
		return
	}
	w.Header().Set("Content-Type", jsonContentType)
	writeBody(w, rootStatusJSON)
}

func scanHandler(client *http.Client, remote bool, limiter *rateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowAnyOrigin(w)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		switch r.Method {
		case http.MethodOptions:
			w.WriteHeader(http.StatusOK)
			return
		case http.MethodPost:
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !limiter.allow(clientKey(r)) {
			http.Error(w, "too many scans, try again in a minute", http.StatusTooManyRequests)
			return
		}
		var req scanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
			http.Error(w, "bad request: missing url", http.StatusBadRequest)
			return
		}
		if !fetch.IsHTTPURL(req.URL) {
			http.Error(w, "bad request: url must start with http:// or https://", http.StatusBadRequest)
			return
		}
		format := report.FormatJSON
		if req.Format != "" {
			f, err := report.ParseFormat(req.Format)
			if err != nil {
				http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
				return
			}
			format = f
		}
		res, err := fetch.Scan(client, req.URL, req.Keyphrase, remote)
		if errors.Is(err, fetch.ErrBlockedAddress) {
			http.Error(w, "bad request: "+fetch.ErrBlockedAddress.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if req.Site && fetch.IsHTTPURL(req.URL) {
			res.Site = fetch.SiteChecks(client, req.URL)
		}
		if format == report.FormatJSON {
			w.Header().Set("Content-Type", jsonContentType)
			_ = json.NewEncoder(w).Encode(res)
			return
		}
		host, _ := report.HostName(req.URL)
		w.Header().Set("Content-Type", formatContentTypes[format])
		writeBody(w, report.PageReport(format, host, audit.Page{URL: req.URL, Scan: res}))
	}
}

func handleScanCount(w http.ResponseWriter, _ *http.Request) {
	allowAnyOrigin(w)
	w.Header().Set("Content-Type", jsonContentType)
	writeBody(w, scanCountJSON)
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	allowAnyOrigin(w)
	writeBody(w, healthJSON)
}
