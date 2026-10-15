package audit

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	bylineClassRegex = regexp.MustCompile(`\b(byline|author)\b`)
	visibleDateRegex = regexp.MustCompile(`(updated|published|modified|posted|reviewed)\s*(on|:)?\s*(([a-z]{3,9}\.? \d{1,2}(st|nd|rd|th)?,? \d{4})|\d{4}-\d{2}-\d{2}|\d{1,2} [a-z]{3,9}\.? \d{4})`)
	experienceRegex  = regexp.MustCompile(`\b(we tested|i tested|we found|i found|we measured|our (research|tests?|study|experience|data|analysis)|in my experience|hands-on|case study)\b`)
	expertiseRegex   = regexp.MustCompile(`\b(ph\.?d|m\.?d|mba|cpa|certified|licensed|board[- ]certified|professor|years of experience|years experience)\b|\bdr\.?\s`)
	reviewedRegex    = regexp.MustCompile(`\b(reviewed by|fact[- ]checked|edited by|medically reviewed|peer[- ]reviewed)\b`)
	referencesRegex  = regexp.MustCompile(`^(sources?|references?|citations?|further reading|bibliography|works cited)\b`)
	maxSnippetRegex  = regexp.MustCompile(`max-snippet:\s*(-?\d+)`)

	authoritativeSuffixes = []string{".gov", ".edu", ".mil", ".ac.uk", ".gov.uk", ".gc.ca"}
	authoritativeDomains  = []string{"wikipedia.org", "who.int", "un.org", "europa.eu", "doi.org", "arxiv.org", "nature.com", "science.org", "sciencedirect.com", "jstor.org", "springer.com", "ieee.org", "acm.org"}
)

func geoSignals(l *checkList, s *PageSignals) {
	geoPublishingDates(l, s)
	geoSourceQuality(l, s)
	geoExperienceSignals(l, s)
	geoSiteTrust(l, s)
}

func metaProperty(s *PageSignals, prop string) string {
	for _, t := range s.Tags["meta"] {
		for _, key := range []string{"property", "name"} {
			if strings.ToLower(strings.TrimSpace(t.Attrs[key])) == prop && strings.TrimSpace(t.Attrs["content"]) != "" {
				return strings.TrimSpace(t.Attrs["content"])
			}
		}
	}
	return ""
}

func schemaDateValue(doc schemaDoc, key string) string {
	for _, node := range doc.nodes {
		if value, ok := node[key].(string); ok && isoDateRegex.MatchString(strings.TrimSpace(value)) {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func geoPublishingDates(l *checkList, s *PageSignals) {
	doc := parseSchemaDoc(s)
	published := schemaDateValue(doc, "datePublished")
	if published == "" {
		published = metaProperty(s, "article:published_time")
	}
	modified := schemaDateValue(doc, "dateModified")
	if modified == "" {
		modified = metaProperty(s, "article:modified_time")
	}
	l.add("geo_date_published", "Publish date declared in schema or article:published_time", published != "")
	l.add("geo_date_modified", "Modified date declared in schema or article:modified_time", modified != "")
	if published == "" || modified == "" || len(published) < 10 || len(modified) < 10 {
		l.skip("geo_modified_after_published", "Needs both a publish and a modified date")
	} else {
		l.add("geo_modified_after_published", "Modified date is not earlier than the publish date", modified[:10] >= published[:10])
	}
	l.add("geo_visible_date", "A visible published or updated date appears on the page", visibleDateRegex.MatchString(strings.ToLower(s.TextContent)))
}

func hostWithoutWWW(host string) string {
	return strings.TrimPrefix(strings.ToLower(host), "www.")
}

func externalHosts(s *PageSignals) []string {
	seen := map[string]bool{}
	var hosts []string
	if s.URL == nil {
		return nil
	}
	own := hostWithoutWWW(s.URL.Host)
	for _, a := range s.Tags["a"] {
		target := resolveNavigable(s, a.Attrs["href"])
		if target == nil || (target.Scheme != "http" && target.Scheme != "https") {
			continue
		}
		if host := hostWithoutWWW(target.Host); host != "" && host != own && !seen[host] {
			seen[host] = true
			hosts = append(hosts, host)
		}
	}
	return hosts
}

func isAuthoritativeHost(host string) bool {
	for _, suffix := range authoritativeSuffixes {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}
	for _, domain := range authoritativeDomains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

func geoSourceQuality(l *checkList, s *PageSignals) {
	hosts := externalHosts(s)
	authoritative := 0
	for _, host := range hosts {
		if isAuthoritativeHost(host) {
			authoritative++
		}
	}
	l.add("geo_authoritative_links", fmt.Sprintf("%d link(s) to government, academic or reference sources", authoritative), authoritative > 0)
	l.add("geo_source_diversity", fmt.Sprintf("%d distinct external source domain(s) (3 or more recommended)", len(hosts)), len(hosts) >= 3)

	citeTags := len(s.Tags["cite"])
	citedQuotes := 0
	for _, name := range []string{"blockquote", "q"} {
		for _, t := range s.Tags[name] {
			if strings.TrimSpace(t.Attrs["cite"]) != "" {
				citedQuotes++
			}
		}
	}
	l.add("geo_cite_markup", "Sources marked up with <cite> or a cite attribute", citeTags+citedQuotes > 0)
	quotes := len(s.Tags["blockquote"])
	skipUnless(l, quotes > 0, "geo_quote_attribution", fmt.Sprintf("%d attributed source(s) for %d blockquote(s)", citeTags+citedQuotes, quotes), "No blockquotes to attribute", citeTags+citedQuotes >= quotes)

	references := false
	for _, b := range contentBlocks(s) {
		references = references || (b.isHeading() && referencesRegex.MatchString(strings.ToLower(b.text)))
	}
	l.add("geo_references_section", "A Sources or References section is present", references)
}

func geoExperienceSignals(l *checkList, s *PageSignals) {
	lower := strings.ToLower(s.TextContent)
	l.add("geo_experience_signals", "First-hand experience wording (we tested, our data, case study)", experienceRegex.MatchString(lower))
	l.add("geo_expertise_signals", "Credentials or experience stated (PhD, certified, years of experience)", expertiseRegex.MatchString(lower))
	l.add("geo_editorial_review", "Reviewed, fact-checked or edited-by statement present", reviewedRegex.MatchString(lower))
	l.add("geo_author_byline", "Visible author byline element (rel=author, itemprop=author or byline class)", hasBylineElement(s))
}

func hasBylineElement(s *PageSignals) bool {
	for name, list := range s.Tags {
		if name == "meta" || name == "link" || name == "html" || name == "head" {
			continue
		}
		for _, t := range list {
			if strings.EqualFold(strings.TrimSpace(t.Attrs["itemprop"]), "author") || bylineClassRegex.MatchString(strings.ToLower(t.Attrs["class"])) {
				return true
			}
			if name == "a" && containsString(strings.Fields(strings.ToLower(t.Attrs["rel"])), "author") {
				return true
			}
		}
	}
	return false
}

func geoSiteTrust(l *checkList, s *PageSignals) {
	policy, llms, contact := false, false, false
	for _, a := range s.Tags["a"] {
		href := strings.ToLower(strings.TrimSpace(a.Attrs["href"]))
		label := href + " " + strings.ToLower(a.Text)
		policy = policy || strings.Contains(label, "privacy") || strings.Contains(label, "terms")
		llms = llms || strings.Contains(href, "llms.txt")
		contact = contact || strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:")
	}
	for _, link := range s.Tags["link"] {
		llms = llms || strings.Contains(strings.ToLower(link.Attrs["href"]), "llms.txt")
	}
	doc := parseSchemaDoc(s)
	for _, node := range doc.nodes {
		contact = contact || jsonPresent(node["email"]) || jsonPresent(node["telephone"]) || jsonPresent(node["contactPoint"])
	}
	l.add("geo_policy_links", "Privacy or terms page linked", policy)
	l.add("geo_llms_txt_link", "llms.txt linked for AI crawlers", llms)
	l.add("geo_contact_details", "Contact details present (mailto, tel, address or schema)", contact || len(s.Tags["address"]) > 0)
	geoSnippetLimit(l, s)
}

func geoSnippetLimit(l *checkList, s *PageSignals) {
	var limits []int
	for _, name := range []string{"robots", "googlebot", "bingbot"} {
		for _, content := range metaTags(s, name) {
			for _, m := range maxSnippetRegex.FindAllStringSubmatch(strings.ToLower(content), -1) {
				if n, err := strconv.Atoi(m[1]); err == nil {
					limits = append(limits, n)
				}
			}
		}
	}
	tooSmall := 0
	for _, n := range limits {
		if n >= 0 && n < 50 {
			tooSmall++
		}
	}
	skipUnless(l, len(limits) > 0, "geo_max_snippet", fmt.Sprintf("%d max-snippet limit(s) below 50 characters", tooSmall), "No max-snippet directive to check", tooSmall == 0)
}
