package fetch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nadimtuhin/seoscore-cli/internal/audit"
)

const (
	userAgent         = "Mozilla/5.0 (compatible; seoscore-cli/1.0; +https://github.com/nadimtuhin/seoscore-cli)"
	defaultRemoteAPI  = "https://seoscore.tools/api/scan"
	remoteAttempts    = 3
	rateLimitBackoff  = 10 * time.Second
	localFileStatus   = http.StatusOK
	fileSchemePrefix  = "file://"
	httpSchemePrefix  = "http://"
	httpsSchemePrefix = "https://"
)

var errStillRateLimited = errors.New("still rate limited after 3 tries")

func IsHTTPURL(target string) bool {
	return strings.HasPrefix(target, httpSchemePrefix) || strings.HasPrefix(target, httpsSchemePrefix)
}

func Scan(client *http.Client, target, keyphrase string, remote bool) (audit.Scan, error) {
	if remote {
		return fetchRemote(client, target, keyphrase)
	}
	if strings.HasPrefix(target, fileSchemePrefix) || !IsHTTPURL(target) {
		return auditLocalFile(strings.TrimPrefix(target, fileSchemePrefix))
	}
	return fetchAndAudit(client, target)
}

func auditLocalFile(filePath string) (audit.Scan, error) {
	content, err := os.ReadFile(filePath) //nolint:gosec // reading the user-named file is the feature; see QUALITY-REPORT.md for the -serve exposure
	if err != nil {
		return audit.Scan{}, fmt.Errorf("read file: %w", err)
	}
	return audit.Analyze(fileSchemePrefix+filepath.Base(filePath), string(content), make(http.Header), localFileStatus), nil
}

func fetchAndAudit(client *http.Client, target string) (audit.Scan, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, http.NoBody)
	if err != nil {
		return audit.Scan{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req) //nolint:gosec // fetching the user-named URL is the feature; see QUALITY-REPORT.md for the -serve exposure
	if err != nil {
		return audit.Scan{}, fmt.Errorf("fetch %s: %w", target, err)
	}
	defer closeBody(resp)

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxPageBytes+1))
	if err != nil {
		return audit.Scan{}, fmt.Errorf("read body: %w", err)
	}
	if len(body) > MaxPageBytes {
		return audit.Scan{}, fmt.Errorf("%s: %w", target, ErrPageTooLarge)
	}
	return audit.Analyze(resp.Request.URL.String(), string(body), resp.Header, resp.StatusCode), nil
}

func closeBody(resp *http.Response) {
	_ = resp.Body.Close()
}

func fetchRemote(client *http.Client, target, keyphrase string) (audit.Scan, error) {
	endpoint := os.Getenv("SEOSCORE_API_URL")
	if endpoint == "" {
		endpoint = defaultRemoteAPI
	}
	reqBody := map[string]string{"url": target}
	if keyphrase != "" {
		reqBody["keyphrase"] = keyphrase
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return audit.Scan{}, fmt.Errorf("encode request: %w", err)
	}
	for range remoteAttempts {
		s, rateLimited, err := postScan(client, endpoint, body)
		if !rateLimited {
			return s, err
		}
		fmt.Fprintln(os.Stderr, "rate limited, waiting 10s")
		time.Sleep(rateLimitBackoff)
	}
	return audit.Scan{}, errStillRateLimited
}

func postScan(client *http.Client, endpoint string, body []byte) (s audit.Scan, rateLimited bool, err error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(body)) //nolint:gosec // endpoint is the operator's own SEOSCORE_API_URL or the default service
	if err != nil {
		return audit.Scan{}, false, fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req) //nolint:gosec // endpoint is the operator's own SEOSCORE_API_URL or the default service
	if err != nil {
		return audit.Scan{}, false, fmt.Errorf("request: %w", err)
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return audit.Scan{}, true, nil
	case http.StatusOK:
	default:
		return audit.Scan{}, false, fmt.Errorf("scan answered %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return audit.Scan{}, false, fmt.Errorf("decode: %w", err)
	}
	return s, false, nil
}

type CrawlOptions struct {
	Keyphrase string
	Limit     int
	Delay     time.Duration
	Remote    bool
	Site      bool
	OnPage    func(done, limit int, page audit.Page)
}

type crawler struct {
	client  *http.Client
	opts    CrawlOptions
	start   *url.URL
	queue   []string
	visited map[string]bool
}

func Crawl(client *http.Client, start string, opts CrawlOptions) ([]audit.Page, error) {
	opts.Limit = max(opts.Limit, 1)
	startURL, _ := url.Parse(start)
	c := &crawler{client: client, opts: opts, start: startURL, queue: []string{start}, visited: map[string]bool{start: true}}
	var pages []audit.Page

	for len(c.queue) > 0 && len(pages) < opts.Limit {
		curr := c.queue[0]
		c.queue = c.queue[1:]

		s, err := scanPage(client, curr, opts)
		if err != nil {
			if len(pages) == 0 {
				return nil, err
			}
			fmt.Fprintf(os.Stderr, "crawl %s: %v\n", curr, err)
			continue
		}

		if len(pages) == 0 {
			c.anchorTo(s.InternalLinks)
		}
		pages = append(pages, audit.Page{URL: curr, Scan: s})
		if opts.OnPage != nil {
			opts.OnPage(len(pages), opts.Limit, pages[len(pages)-1])
		}
		c.enqueue(s.InternalLinks)

		if len(c.queue) > 0 && len(pages) < opts.Limit && opts.Delay > 0 && opts.Remote {
			time.Sleep(opts.Delay)
		}
	}
	return pages, nil
}

func (c *crawler) anchorTo(links []string) {
	for _, link := range links {
		u, err := url.Parse(link)
		if err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != "" {
			c.start = u
			return
		}
	}
}

func (c *crawler) enqueue(links []string) {
	for _, link := range links {
		lu, err := url.Parse(link)
		if err != nil || (lu.Scheme != "http" && lu.Scheme != "https") {
			continue
		}
		if c.start != nil && lu.Hostname() != c.start.Hostname() {
			continue
		}
		lu.Fragment = ""
		clean := lu.String()
		if !c.visited[clean] {
			c.visited[clean] = true
			c.queue = append(c.queue, clean)
		}
	}
}
