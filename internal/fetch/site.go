package fetch

import (
	"net/http"
	"net/url"

	"github.com/nadimtuhin/seoscore-cli/internal/audit"
)

func SiteChecks(client *http.Client, pageURL string) *audit.Category {
	u, err := url.Parse(pageURL)
	if err != nil {
		cat := audit.SiteCategory("/", nil, nil)
		return &cat
	}
	origin := u.Scheme + "://" + u.Host
	cat := audit.SiteCategory(u.RequestURI(), optionalBody(client, origin+"/robots.txt"), optionalBody(client, origin+"/llms.txt"))
	return &cat
}

func optionalBody(client *http.Client, target string) *string {
	body, err := getBody(client, target)
	if err != nil {
		return nil
	}
	text := string(body)
	return &text
}

func scanPage(client *http.Client, target string, opts CrawlOptions) (audit.Scan, error) {
	s, err := Scan(client, target, opts.Keyphrase, opts.Remote)
	if err == nil && opts.Site && IsHTTPURL(target) {
		s.Site = SiteChecks(client, target)
	}
	return s, err
}
