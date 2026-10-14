package fetch

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nadimtuhin/seoscore-cli/internal/audit"
)

const (
	maxSitemapBytes = 10 << 20
	maxSitemapDepth = 3
)

type sitemapDoc struct {
	XMLName  xml.Name
	URLs     []string `xml:"url>loc"`
	Sitemaps []string `xml:"sitemap>loc"`
}

func parseSitemap(body []byte) (pages, children []string, err error) {
	var doc sitemapDoc
	if err := xml.Unmarshal(body, &doc); err != nil {
		return nil, nil, fmt.Errorf("parse sitemap: %w", err)
	}
	return trimAll(doc.URLs), trimAll(doc.Sitemaps), nil
}

func trimAll(locs []string) []string {
	out := make([]string, 0, len(locs))
	for _, l := range locs {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

type sitemapWalk struct {
	client  *http.Client
	limit   int
	seen    map[string]bool
	visited map[string]bool
	urls    []string
}

func SitemapURLs(client *http.Client, sitemapURL string, limit int) ([]string, error) {
	w := &sitemapWalk{client: client, limit: max(limit, 1), seen: map[string]bool{}, visited: map[string]bool{}}
	if err := w.visit(sitemapURL, 0); err != nil {
		return nil, err
	}
	return w.urls, nil
}

func (w *sitemapWalk) visit(sitemapURL string, depth int) error {
	if w.visited[sitemapURL] {
		return nil
	}
	w.visited[sitemapURL] = true
	body, err := getBody(w.client, sitemapURL)
	if err != nil {
		return err
	}
	pages, children, err := parseSitemap(body)
	if err != nil {
		return fmt.Errorf("%s: %w", sitemapURL, err)
	}
	for _, p := range pages {
		if len(w.urls) == w.limit {
			return nil
		}
		if !w.seen[p] {
			w.seen[p] = true
			w.urls = append(w.urls, p)
		}
	}
	if depth == maxSitemapDepth {
		return nil
	}
	for _, child := range children {
		if len(w.urls) == w.limit {
			return nil
		}
		if err := w.visit(child, depth+1); err != nil {
			fmt.Fprintf(os.Stderr, "sitemap %s: %v\n", child, err)
		}
	}
	return nil
}

func getBody(client *http.Client, target string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req) //nolint:gosec // fetching the user-named sitemap is the feature
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", target, err)
	}
	defer closeBody(resp)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: %s", target, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSitemapBytes))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", target, err)
	}
	return body, nil
}

func ScanAll(client *http.Client, targets []string, opts CrawlOptions) ([]audit.Page, error) {
	var pages []audit.Page
	for i, target := range targets {
		s, err := scanPage(client, target, opts)
		if err != nil {
			if len(pages) == 0 && i == len(targets)-1 {
				return nil, fmt.Errorf("all %d pages failed, last: %w", len(targets), err)
			}
			fmt.Fprintf(os.Stderr, "scan %s: %v\n", target, err)
			continue
		}
		pages = append(pages, audit.Page{URL: target, Scan: s})
		if opts.OnPage != nil {
			opts.OnPage(len(pages), len(targets), pages[len(pages)-1])
		}
		if i < len(targets)-1 && opts.Delay > 0 && opts.Remote {
			time.Sleep(opts.Delay)
		}
	}
	return pages, nil
}
