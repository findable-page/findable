package main

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

func isHTTPURL(target string) bool {
	return strings.HasPrefix(target, httpSchemePrefix) || strings.HasPrefix(target, httpsSchemePrefix)
}

func scanTarget(client *http.Client, target, keyphrase string, remote bool) (scan, error) {
	if remote {
		return fetchRemote(client, target, keyphrase)
	}
	if strings.HasPrefix(target, fileSchemePrefix) || !isHTTPURL(target) {
		return auditLocalFile(strings.TrimPrefix(target, fileSchemePrefix))
	}
	return fetchAndAudit(client, target)
}

func auditLocalFile(filePath string) (scan, error) {
	content, err := os.ReadFile(filePath) //nolint:gosec // reading the user-named file is the feature; see QUALITY-REPORT.md for the -serve exposure
	if err != nil {
		return scan{}, fmt.Errorf("read file: %w", err)
	}
	return AuditPage(fileSchemePrefix+filepath.Base(filePath), string(content), make(http.Header), localFileStatus), nil
}

func fetchAndAudit(client *http.Client, target string) (scan, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, http.NoBody)
	if err != nil {
		return scan{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req) //nolint:gosec // fetching the user-named URL is the feature; see QUALITY-REPORT.md for the -serve exposure
	if err != nil {
		return scan{}, fmt.Errorf("fetch %s: %w", target, err)
	}
	defer closeBody(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return scan{}, fmt.Errorf("read body: %w", err)
	}
	return AuditPage(target, string(body), resp.Header, resp.StatusCode), nil
}

func closeBody(resp *http.Response) {
	_ = resp.Body.Close()
}

func fetchRemote(client *http.Client, target, keyphrase string) (scan, error) {
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
		return scan{}, fmt.Errorf("encode request: %w", err)
	}
	for range remoteAttempts {
		s, rateLimited, err := postScan(client, endpoint, body)
		if !rateLimited {
			return s, err
		}
		fmt.Fprintln(os.Stderr, "rate limited, waiting 10s")
		time.Sleep(rateLimitBackoff)
	}
	return scan{}, errStillRateLimited
}

func postScan(client *http.Client, endpoint string, body []byte) (s scan, rateLimited bool, err error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(body)) //nolint:gosec // endpoint is the operator's own SEOSCORE_API_URL or the default service
	if err != nil {
		return scan{}, false, fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req) //nolint:gosec // endpoint is the operator's own SEOSCORE_API_URL or the default service
	if err != nil {
		return scan{}, false, fmt.Errorf("request: %w", err)
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return scan{}, true, nil
	case http.StatusOK:
	default:
		return scan{}, false, fmt.Errorf("scan answered %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return scan{}, false, fmt.Errorf("decode: %w", err)
	}
	return s, false, nil
}

type crawlOptions struct {
	keyphrase string
	limit     int
	delay     time.Duration
	remote    bool
	onPage    func(done, limit int, page crawledPage)
}

type crawler struct {
	client  *http.Client
	opts    crawlOptions
	start   *url.URL
	queue   []string
	visited map[string]bool
}

func crawl(client *http.Client, start string, opts crawlOptions) ([]crawledPage, error) {
	opts.limit = max(opts.limit, 1)
	startURL, _ := url.Parse(start)
	c := &crawler{client: client, opts: opts, start: startURL, queue: []string{start}, visited: map[string]bool{start: true}}
	var pages []crawledPage

	for len(c.queue) > 0 && len(pages) < opts.limit {
		curr := c.queue[0]
		c.queue = c.queue[1:]

		s, err := scanTarget(client, curr, opts.keyphrase, opts.remote)
		if err != nil {
			if len(pages) == 0 {
				return nil, err
			}
			fmt.Fprintf(os.Stderr, "crawl %s: %v\n", curr, err)
			continue
		}

		pages = append(pages, crawledPage{url: curr, scan: s})
		if opts.onPage != nil {
			opts.onPage(len(pages), opts.limit, pages[len(pages)-1])
		}
		c.enqueue(s.InternalLinks)

		if len(c.queue) > 0 && len(pages) < opts.limit && opts.delay > 0 && opts.remote {
			time.Sleep(opts.delay)
		}
	}
	return pages, nil
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
