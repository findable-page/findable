package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/findable-page/findable/internal/audit"
)

const (
	userAgent         = "Mozilla/5.0 (compatible; findable/1.0; +https://github.com/findable-page/findable)"
	localFileStatus   = http.StatusOK
	fileSchemePrefix  = "file://"
	httpSchemePrefix  = "http://"
	httpsSchemePrefix = "https://"
)

func IsHTTPURL(target string) bool {
	return strings.HasPrefix(target, httpSchemePrefix) || strings.HasPrefix(target, httpsSchemePrefix)
}

func Scan(client *http.Client, target, keyphrase string) (audit.Scan, error) {
	if strings.HasPrefix(target, fileSchemePrefix) || !IsHTTPURL(target) {
		return auditLocalFile(strings.TrimPrefix(target, fileSchemePrefix), keyphrase)
	}
	return fetchAndAudit(client, target, keyphrase)
}

func auditLocalFile(filePath, keyphrase string) (audit.Scan, error) {
	content, err := os.ReadFile(filePath) //nolint:gosec // reading the user-named file is the feature; see the SSRF policy in SECURITY.md
	if err != nil {
		return audit.Scan{}, fmt.Errorf("read file: %w", err)
	}
	return audit.Analyze(fileSchemePrefix+filepath.Base(filePath), string(content), keyphrase, make(http.Header), localFileStatus), nil
}

func fetchAndAudit(client *http.Client, target, keyphrase string) (audit.Scan, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, http.NoBody)
	if err != nil {
		return audit.Scan{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req) //nolint:gosec // fetching the user-named URL is the feature; see the SSRF policy in SECURITY.md
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
	return audit.Analyze(resp.Request.URL.String(), string(body), keyphrase, resp.Header, resp.StatusCode), nil
}

func closeBody(resp *http.Response) {
	_ = resp.Body.Close()
}

type CrawlOptions struct {
	Keyphrase string
	Limit     int
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
