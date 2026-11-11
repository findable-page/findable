package audit

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	visibleDateRegex = regexp.MustCompile(`(updated|published|modified|posted|reviewed)\s*(on|:)?\s*(([a-z]{3,9}\.? \d{1,2}(st|nd|rd|th)?,? \d{4})|\d{4}-\d{2}-\d{2}|\d{1,2} [a-z]{3,9}\.? \d{4})`)

	authoritativeSuffixes = []string{".gov", ".edu", ".mil", ".ac.uk", ".gov.uk", ".gc.ca"}
	authoritativeDomains  = []string{"wikipedia.org", "who.int", "un.org", "europa.eu", "doi.org", "arxiv.org", "nature.com", "science.org", "sciencedirect.com", "jstor.org", "springer.com", "ieee.org", "acm.org"}
)

func geoSignals(l *checkList, s *PageSignals) {
	doc := parseSchemaDoc(s)
	article := isArticlePage(s, doc)
	geoPublishingDates(l, s, article)
	geoSourceQuality(l, s, article)
	geoSiteTrust(l, s)
	geoSchemaTopicalDepth(l, doc, article)
}

func isArticlePage(s *PageSignals, doc schemaDoc) bool {
	return len(doc.ofType(articleTypes...)) > 0 || metaProperty(s, "article:published_time") != "" || articleContainsHeading(s)
}

func articleContainsHeading(s *PageSignals) bool {
	if len(s.Tags["article"]) != 1 {
		return false
	}
	lower := strings.ToLower(s.RawHTML)
	start := strings.Index(lower, "<article")
	if start == -1 {
		return false
	}
	end := len(lower)
	if closeAt := strings.Index(lower[start:], "</article"); closeAt != -1 {
		end = start + closeAt
	}
	h1 := strings.Index(lower[start:end], "<h1")
	return h1 != -1
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

func geoPublishingDates(l *checkList, s *PageSignals, article bool) {
	doc := parseSchemaDoc(s)
	published := schemaDateValue(doc, "datePublished")
	if published == "" {
		published = metaProperty(s, "article:published_time")
	}
	modified := schemaDateValue(doc, "dateModified")
	if modified == "" {
		modified = metaProperty(s, "article:modified_time")
	}
	notArticle := "Not an article page"
	skipUnless(l, article, "geo_date_published", "Publish date declared in schema or article:published_time", notArticle, published != "")
	skipUnless(l, article, "geo_date_modified", "Modified date declared in schema or article:modified_time", notArticle, modified != "")
	if published == "" || modified == "" || len(published) < 10 || len(modified) < 10 {
		l.skip("geo_modified_after_published", "Needs both a publish and a modified date")
	} else {
		l.add("geo_modified_after_published", "Modified date is not earlier than the publish date", modified[:10] >= published[:10])
	}
	skipUnless(l, article, "geo_visible_date", "A visible published or updated date appears on the page", notArticle, visibleDateRegex.MatchString(strings.ToLower(s.TextContent)))
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
		if target == nil || (target.Scheme != "http" && target.Scheme != schemeHTTPS) {
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

func geoSourceQuality(l *checkList, s *PageSignals, article bool) {
	hosts := externalHosts(s)
	authoritative := 0
	for _, host := range hosts {
		if isAuthoritativeHost(host) {
			authoritative++
		}
	}
	skipUnless(l, article, "geo_authoritative_links", fmt.Sprintf("%d link(s) to government, academic or reference sources", authoritative), "Not an article page", authoritative > 0)
	skipUnless(l, article, "geo_source_diversity", fmt.Sprintf("%d distinct external source domain(s) (3 or more recommended)", len(hosts)), "Not an article page", len(hosts) >= 3)

	citeTags := len(s.Tags["cite"])
	citedQuotes := 0
	for _, name := range []string{"blockquote", "q"} {
		for _, t := range s.Tags[name] {
			if strings.TrimSpace(t.Attrs["cite"]) != "" {
				citedQuotes++
			}
		}
	}
	quotes := len(s.Tags["blockquote"])
	skipUnless(l, quotes > 0, "geo_quote_attribution", fmt.Sprintf("%d attributed source(s) for %d blockquote(s)", citeTags+citedQuotes, quotes), "No blockquotes to attribute", citeTags+citedQuotes >= quotes)
}

func geoSiteTrust(l *checkList, s *PageSignals) {
	policy := false
	for _, a := range s.Tags["a"] {
		label := strings.ToLower(strings.TrimSpace(a.Attrs["href"])) + " " + strings.ToLower(a.Text)
		policy = policy || strings.Contains(label, "privacy") || strings.Contains(label, "terms")
	}
	l.add("geo_policy_links", "Privacy or terms page linked", policy)
}

const knowsAboutMinTopics = 3

var credentialProfileHosts = []string{
	"linkedin.com", "orcid.org", "researchgate.net", "scholar.google.com", "crunchbase.com", "github.com",
}

// geoSchemaTopicalDepth covers the schema-topical-eeat-depth capability:
// organizations that enumerate their topical competence in knowsAbout, and
// article authors whose Person schema carries verifiable credentials.
func geoSchemaTopicalDepth(l *checkList, doc schemaDoc, article bool) {
	geoSchemaKnowsAbout(l, doc)
	geoAuthorCredentialDepth(l, doc, article)
}

// geoSchemaKnowsAbout rewards Organization or WebSite nodes declaring at
// least knowsAboutMinTopics topics. Entries may be plain strings or Thing
// objects carrying a name or @id (schema.org allows both).
func geoSchemaKnowsAbout(l *checkList, doc schemaDoc) {
	nodes := append(doc.ofType(append(append([]string{}, orgTypes...), localTypes...)...), doc.ofType("website")...)
	best := 0
	for _, n := range nodes {
		if count := knowsAboutCount(n); count > best {
			best = count
		}
	}
	skipUnless(l, len(nodes) > 0, "geo_schema_knows_about",
		fmt.Sprintf("Organization or WebSite declares %d topical area(s) in knowsAbout (%d or more recommended)", best, knowsAboutMinTopics),
		"No Organization or WebSite schema to validate",
		best >= knowsAboutMinTopics)
}

func knowsAboutCount(n map[string]any) int {
	count := 0
	for _, item := range jsonItems(n["knowsAbout"]) {
		if m, ok := item.(map[string]any); ok {
			if jsonPresent(m["name"]) || jsonPresent(m["@id"]) {
				count++
			}
			continue
		}
		if jsonPresent(item) {
			count++
		}
	}
	return count
}

// geoAuthorCredentialDepth verifies that the Person schema behind article
// authors carries at least one verifiable qualification: jobTitle, alumniOf,
// hasCredential, or a professional profile among the sameAs links.
func geoAuthorCredentialDepth(l *checkList, doc schemaDoc, article bool) {
	persons := authorPersons(doc)
	skipUnless(l, article && len(persons) > 0, "geo_author_credential_depth",
		"Author schema includes professional credentials (jobTitle, alumniOf, hasCredential, or profile sameAs)",
		"No article author Person schema to validate",
		allNodes(persons, personHasCredential))
}

// authorPersons returns root-level Person nodes plus Person author objects
// nested inside Article nodes, which parseSchemaDoc does not collect on its
// own.
func authorPersons(doc schemaDoc) []map[string]any {
	persons := doc.ofType("person")
	for _, a := range doc.ofType(articleTypes...) {
		for _, item := range jsonItems(a["author"]) {
			if m, ok := item.(map[string]any); ok && containsString(nodeTypes(m), "person") {
				persons = append(persons, m)
			}
		}
	}
	return persons
}

func personHasCredential(n map[string]any) bool {
	for _, key := range []string{"jobTitle", "alumniOf", "hasCredential"} {
		if jsonPresent(n[key]) {
			return true
		}
	}
	for _, link := range sameAsLinks(n["sameAs"]) {
		lower := strings.ToLower(link)
		for _, host := range credentialProfileHosts {
			if strings.Contains(lower, host) {
				return true
			}
		}
	}
	return false
}
