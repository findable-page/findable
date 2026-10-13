// seoscore scans a site locally or via seoscore.tools and saves the report CSV.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type check struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Pass       bool   `json:"pass"`
	Applicable *bool  `json:"applicable"`
}

type category struct {
	Score  *float64 `json:"score"`
	Passed int      `json:"passed"`
	Total  int      `json:"total"`
	Checks []check  `json:"checks"`
}

type scan struct {
	SEO           category `json:"seo"`
	AEO           category `json:"aeo"`
	GEO           category `json:"geo"`
	CWV           category `json:"cwv"`
	InternalLinks []string `json:"internalLinks"`
}

type named struct {
	name string
	cat  category
}

func (s scan) scored() []named {
	return []named{{"SEO", s.SEO}, {"AEO", s.AEO}, {"GEO", s.GEO}}
}

type crawledPage struct {
	url  string
	scan scan
}

func main() {
	outDir := flag.String("o", ".", "directory for the CSV")
	quiet := flag.Bool("q", false, "do not list failing checks")
	keyphrase := flag.String("k", "", "target keyphrase for audit")
	crawl := flag.Bool("crawl", false, "crawl internal links up to limit pages")
	limit := flag.Int("limit", 20, "max pages to crawl when -crawl is enabled")
	delay := flag.Duration("delay", 3*time.Second, "delay between crawl requests")
	remote := flag.Bool("remote", false, "use remote seoscore.tools API instead of local engine")
	serve := flag.String("serve", "", "run HTTP API server on [host]:port (e.g. :8080)")
	mcpMode := flag.Bool("mcp", false, "run Model Context Protocol (MCP) server over stdio")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: seoscore [-o dir] [-q] [-k keyphrase] [-crawl] [-limit N] [-remote] [-serve :port] [-mcp] <url or file>")
	}
	flag.Parse()

	client := &http.Client{Timeout: 60 * time.Second}

	if *mcpMode {
		runMCP(os.Stdin, os.Stdout, client)
		return
	}

	if *serve != "" {
		startServer(*serve, client, *remote)
		return
	}

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	target := flag.Arg(0)
	isLocalFile := !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://")

	var hostName string
	if isLocalFile {
		hostName = filepath.Base(target)
	} else {
		u, err := url.Parse(target)
		if err != nil || u.Host == "" {
			fail("need a valid URL or local HTML file path, got %q", target)
		}
		hostName = u.Host
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fail("create %s: %v", *outDir, err)
	}

	if *crawl {
		if isLocalFile {
			fail("-crawl requires an http(s) URL, not a local file")
		}
		runCrawl(client, target, hostName, *outDir, *keyphrase, *limit, *delay, *quiet, *remote)
		return
	}

	s, err := scanTarget(client, target, *keyphrase, *remote)
	if err != nil {
		fail("%v", err)
	}

	path := filepath.Join(*outDir, fmt.Sprintf("seoscore-report-%s-%s.csv", hostName, time.Now().Format("2006-01-02")))
	if err := os.WriteFile(path, []byte(csv(hostName, s)), 0o644); err != nil {
		fail("write %s: %v", path, err)
	}

	fmt.Printf("overall %d  SEO %s  AEO %s  GEO %s\n", overall(s), score(s.SEO), score(s.AEO), score(s.GEO))
	if !*quiet {
		for _, c := range s.scored() {
			for _, k := range c.cat.Checks {
				if status(k) == "FAIL" {
					fmt.Printf("FAIL %s %s: %s\n", c.name, k.ID, k.Label)
				}
			}
		}
	}
	fmt.Println("saved", path)
}

func scanTarget(client *http.Client, target, keyphrase string, remote bool) (scan, error) {
	if remote {
		return fetchRemote(client, target, keyphrase)
	}
	return fetchLocal(client, target)
}

func fetchLocal(client *http.Client, target string) (scan, error) {
	if strings.HasPrefix(target, "file://") || (!strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://")) {
		filePath := strings.TrimPrefix(target, "file://")
		content, err := os.ReadFile(filePath)
		if err != nil {
			return scan{}, fmt.Errorf("read file: %w", err)
		}
		return AuditPage("file://"+filepath.Base(filePath), string(content), make(http.Header), 200), nil
	}

	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return scan{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; seoscore-cli/1.0; +https://github.com/nadimtuhin/seoscore-cli)")

	resp, err := client.Do(req)
	if err != nil {
		return scan{}, fmt.Errorf("fetch %s: %w", target, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return scan{}, fmt.Errorf("read body: %w", err)
	}

	return AuditPage(target, string(bodyBytes), resp.Header, resp.StatusCode), nil
}

func fetchRemote(client *http.Client, target, keyphrase string) (scan, error) {
	apiEndpoint := os.Getenv("SEOSCORE_API_URL")
	if apiEndpoint == "" {
		apiEndpoint = "https://seoscore.tools/api/scan"
	}
	reqBody := map[string]string{"url": target}
	if keyphrase != "" {
		reqBody["keyphrase"] = keyphrase
	}
	body, _ := json.Marshal(reqBody)
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := client.Post(apiEndpoint, "application/json", bytes.NewReader(body))
		if err != nil {
			return scan{}, fmt.Errorf("request: %w", err)
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			fmt.Fprintln(os.Stderr, "rate limited, waiting 10s")
			time.Sleep(10 * time.Second)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return scan{}, fmt.Errorf("scan answered %s", resp.Status)
		}
		var s scan
		err = json.NewDecoder(resp.Body).Decode(&s)
		resp.Body.Close()
		if err != nil {
			return scan{}, fmt.Errorf("decode: %w", err)
		}
		return s, nil
	}
	return scan{}, fmt.Errorf("still rate limited after 3 tries")
}

func runCrawl(client *http.Client, start string, hostName string, outDir, keyphrase string, limit int, delay time.Duration, quiet bool, remote bool) {
	if limit < 1 {
		limit = 1
	}
	startURL, _ := url.Parse(start)
	queue := []string{start}
	visited := map[string]bool{start: true}
	var pages []crawledPage

	for len(queue) > 0 && len(pages) < limit {
		curr := queue[0]
		queue = queue[1:]

		s, err := scanTarget(client, curr, keyphrase, remote)
		if err != nil {
			if len(pages) == 0 {
				fail("%v", err)
			}
			fmt.Fprintf(os.Stderr, "crawl %s: %v\n", curr, err)
			continue
		}

		pages = append(pages, crawledPage{url: curr, scan: s})
		fmt.Printf("[%d/%d] %s  overall %d  SEO %s  AEO %s  GEO %s  failed %d\n",
			len(pages), limit, curr, overall(s), score(s.SEO), score(s.AEO), score(s.GEO), failedCount(s))

		if !quiet {
			for _, c := range s.scored() {
				for _, k := range c.cat.Checks {
					if status(k) == "FAIL" {
						fmt.Printf("  FAIL %s %s: %s\n", c.name, k.ID, k.Label)
					}
				}
			}
		}

		for _, link := range s.InternalLinks {
			lu, err := url.Parse(link)
			if err != nil || (lu.Scheme != "http" && lu.Scheme != "https") {
				continue
			}
			if startURL != nil && lu.Hostname() != startURL.Hostname() {
				continue
			}
			lu.Fragment = ""
			clean := lu.String()
			if !visited[clean] {
				visited[clean] = true
				queue = append(queue, clean)
			}
		}

		if len(queue) > 0 && len(pages) < limit && delay > 0 && remote {
			time.Sleep(delay)
		}
	}

	path := filepath.Join(outDir, fmt.Sprintf("seoscore-site-report-%s-%s.csv", hostName, time.Now().Format("2006-01-02")))
	if err := os.WriteFile(path, []byte(crawlCSV(hostName, pages)), 0o644); err != nil {
		fail("write %s: %v", path, err)
	}
	fmt.Printf("saved %s (%d pages)\n", path, len(pages))
}

func startServer(addr string, client *http.Client, remote bool) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if strings.Contains(r.Header.Get("Accept"), "text/html") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(localWebUIHTML))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","name":"seoscore-cli-server"}`))
	})
	mux.HandleFunc("/api/scan", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			URL       string `json:"url"`
			Keyphrase string `json:"keyphrase"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
			http.Error(w, "bad request: missing url", http.StatusBadRequest)
			return
		}
		res, err := scanTarget(client, req.URL, req.Keyphrase, remote)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})
	mux.HandleFunc("/api/scan-count", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":2847}`))
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(`{"status":"ok"}`))
	})
	fmt.Printf("seoscore API server listening on %s (remote=%v)\n", addr, remote)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fail("server error: %v", err)
	}
}

func status(c check) string {
	switch {
	case c.Applicable != nil && !*c.Applicable:
		return "N/A"
	case c.Pass:
		return "PASS"
	}
	return "FAIL"
}

func val(c category) float64 {
	if c.Score == nil {
		return 0
	}
	return *c.Score
}

func score(c category) string { return fmt.Sprintf("%.0f", val(c)) }

func overall(s scan) int {
	return int(math.Round(val(s.SEO)*.5 + val(s.AEO)*.25 + val(s.GEO)*.25))
}

func failedCount(s scan) int {
	cnt := 0
	for _, c := range s.scored() {
		for _, k := range c.cat.Checks {
			if status(k) == "FAIL" {
				cnt++
			}
		}
	}
	return cnt
}

func quote(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }

func csv(host string, s scan) string {
	var b strings.Builder
	row := func(cells ...string) { b.WriteString(strings.Join(cells, ",") + "\n") }
	b.WriteString("\xef\xbb\xbf")
	row("SEO/AEO/GEO Report", host, time.Now().Format("2006-01-02"))
	row("Generated by", "seoscore.tools", "")
	row("")
	row("Category", "Score", "Passed", "Total")
	row("Overall", fmt.Sprint(overall(s)), "", "")
	for _, c := range s.scored() {
		row(c.name, score(c.cat), fmt.Sprint(c.cat.Passed), fmt.Sprint(c.cat.Total))
	}
	if s.CWV.Score != nil {
		row("CWV", score(s.CWV), fmt.Sprint(s.CWV.Passed), fmt.Sprint(s.CWV.Total))
	}
	row("")
	row("Check ID", "Category", "Status", "Description")
	for _, c := range append(s.scored(), named{"CWV", s.CWV}) {
		for _, k := range c.cat.Checks {
			row(k.ID, c.name, status(k), quote(k.Label))
		}
	}
	return b.String()
}

func crawlCSV(host string, pages []crawledPage) string {
	var b strings.Builder
	row := func(cells ...string) { b.WriteString(strings.Join(cells, ",") + "\n") }
	b.WriteString("\xef\xbb\xbf")
	date := time.Now().Format("2006-01-02")
	row("Site Crawl Report", host, date)
	row("Generated by", "seoscore.tools", "")
	row("Pages scanned", fmt.Sprint(len(pages)), "")
	row("")
	row("URL", "SEO Score", "AEO Score", "GEO Score", "Overall", "Failed Checks")
	for _, p := range pages {
		row(quote(p.url), score(p.scan.SEO), score(p.scan.AEO), score(p.scan.GEO), fmt.Sprint(overall(p.scan)), fmt.Sprint(failedCount(p.scan)))
	}
	row("")
	row("All Failed Checks", "", "", "", "", "")
	row("URL", "Check ID", "Category", "Description")
	for _, p := range pages {
		for _, c := range p.scan.scored() {
			for _, k := range c.cat.Checks {
				if status(k) == "FAIL" {
					row(quote(p.url), k.ID, c.name, quote(k.Label))
				}
			}
		}
	}
	return b.String()
}

func fail(f string, a ...any) {
	fmt.Fprintf(os.Stderr, "seoscore: "+f+"\n", a...)
	os.Exit(1)
}
