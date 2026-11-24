package audit

import (
	"fmt"
	"net/url"
	"strings"
)

func evalGEO(s *PageSignals) []Check {
	return runChecks(s,
		geoData,
		geoCitations,
		geoAuthority,
		geoTrust,
		geoFreshness,
		geoIntent,
		geoSignals,
	)
}

func geoData(l *checkList, s *PageSignals) {
	stats := len(s.NumbersWithUnits)
	l.add("geo_statistics_density", fmt.Sprintf("%d quantitative data points / metrics found", stats), stats >= 3)
	l.add("geo_data_viz", "Structured data tables or charts found", s.TableCount > 0)
	l.add("geo_comparison", "Comparative sections or tables detected", s.TableCount > 0 || lowerContainsAny(s.TextContent, " vs "))
}

func geoCitations(l *checkList, s *PageSignals) {
	l.add("geo_expert_quotes", fmt.Sprintf("%d expert quotes/blockquotes found", len(s.ExpertQuotes)), len(s.ExpertQuotes) > 0)
	l.add("geo_citations", fmt.Sprintf("%d external source citations found", len(s.ExternalLinks)), len(s.ExternalLinks) > 0)
	l.add("geo_definitions", "Clear concept definitions found ('X is a...')", lowerContainsAny(s.TextContent, " is a ", " is the ", " defined as "))
	l.add("geo_passage_citability", "Passage-ready standalone answer blocks present", hasParagraphOfLength(s.Paragraphs, 40, 120))
}

func geoAuthority(l *checkList, s *PageSignals) {
	hasPerson := hasSchemaType(s.SchemaTypes, "Person")
	hasOrg := hasSchemaType(s.SchemaTypes, "Organization", "Bakery", "LocalBusiness")
	l.add("geo_author_schema", "Author schema markup found", hasPerson || s.AuthorInSchema != "")
	l.add("geo_author_credentials", "Visible author byline with credentials or dates", lowerContainsAny(s.TextContent, "by ") || s.AuthorInSchema != "")
	l.add("geo_org_schema", "Organization or Business schema markup declared", hasOrg)
	l.add("geo_knowledge_graph", fmt.Sprintf("%d sameAs entity links in schema", len(s.SameAsLinks)), len(s.SameAsLinks) > 0)
	geoEntityAuthoritativeSameAs(l, s)
}

// authoritativeEntityDomains are the entity registries answer engines use to
// resolve entities: a sameAs link to one of them anchors the Organization or
// Person node in the knowledge graph.
var authoritativeEntityDomains = []string{"wikidata.org", "wikipedia.org", "crunchbase.com"}

// geoEntityAuthoritativeSameAs verifies that schema sameAs links point to
// authoritative entity repositories rather than only to the site's own pages
// or social profiles. Skipped when no sameAs links exist at all because
// geo_knowledge_graph already covers that case.
func geoEntityAuthoritativeSameAs(l *checkList, s *PageSignals) {
	if len(s.SameAsLinks) == 0 {
		l.skip("geo_entity_authoritative_sameas", "No sameAs links in schema to evaluate")
		return
	}
	count := 0
	var ev []Evidence
	for _, link := range s.SameAsLinks {
		if isAuthoritativeEntityURL(link) {
			count++
		} else if len(ev) < maxEvidence {
			ev = append(ev, evidence("sameAs", link))
		}
	}
	l.addEv("geo_entity_authoritative_sameas",
		fmt.Sprintf("%d of %d sameAs links point to authoritative entity registries (Wikidata, Wikipedia, Crunchbase)", count, len(s.SameAsLinks)),
		count > 0, ev...)
}

// isAuthoritativeEntityURL reports whether the URL's host is one of the
// authoritative entity registry domains or a subdomain of one (language
// wikis like en.wikipedia.org included).
func isAuthoritativeEntityURL(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, domain := range authoritativeEntityDomains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

func geoTrust(l *checkList, s *PageSignals) {
	hasAbout := false
	hasContact := false
	for _, link := range s.InternalLinks {
		hasAbout = hasAbout || lowerContainsAny(link, "about")
		hasContact = hasContact || lowerContainsAny(link, "contact")
	}
	l.add("geo_trust", "Trust signals (About and Contact links) present", hasAbout && hasContact)
	l.add("geo_entity_markup", fmt.Sprintf("%d schema entity type(s) declared", len(s.SchemaTypes)), len(s.SchemaTypes) >= 2)
}

func geoFreshness(l *checkList, s *PageSignals) {
	hasTime := s.SemanticCounts["time"] > 0
	l.add("geo_freshness", "Content freshness signals (dates/time elements) present", hasTime || yearRegex.MatchString(s.TextContent) || s.DateModified != "")
	l.add("geo_time_element", "<time> machine-readable tags present", hasTime)
	l.add("geo_content_length", fmt.Sprintf("%d words content depth (300+ recommended)", len(s.Words)), len(s.Words) >= 300)
}

func geoIntent(l *checkList, s *PageSignals) {
	h1 := firstH1Text(s.Headings)
	skipUnless(l, s.Title != "" && h1 != "", "geo_intent_clarity", "Title keywords match primary H1 heading", "No title or H1 heading to compare", titleMatchesHeading(s.Title, h1))
	formats := formattingTypeCount(s)
	l.add("geo_format_variety", fmt.Sprintf("%d formatting types used (tables, lists, quotes, code, dl)", formats), formats >= 3)
	l.add("geo_snippet_control", "No restrictive nosnippet over-blocking", !lowerContainsAny(s.RobotsMeta, "nosnippet"))
	geoSnippetBoilerplateControl(l, s)
}

const boilerplateWordThreshold = 30

// geoSnippetBoilerplateControl flags long navigation, header and footer text
// that AI engines can lift into snippets. data-nosnippet only counts on div,
// span and section elements, so the attribute on the nav or footer itself is
// ignored. A block passes when most of its words sit under data-nosnippet
// wrappers.
func geoSnippetBoilerplateControl(l *checkList, s *PageSignals) {
	significant, controlled := 0, 0
	for _, b := range s.Boilerplate {
		if b.Words < boilerplateWordThreshold {
			continue
		}
		significant++
		if b.CoveredWords*2 >= b.Words {
			controlled++
		}
	}
	if significant == 0 {
		l.skip("geo_snippet_boilerplate_control", "No long nav, header or footer text to scope")
		return
	}
	l.add("geo_snippet_boilerplate_control", fmt.Sprintf("%d of %d long nav/header/footer block(s) mostly wrapped in data-nosnippet div/span/section", controlled, significant), controlled == significant)
}

func firstH1Text(headings []Heading) string {
	for _, h := range headings {
		if h.Level == 1 {
			return h.Text
		}
	}
	return ""
}

func titleMatchesHeading(title, h1 string) bool {
	firstHeading := strings.ToLower(h1)
	matches := 0
	for _, word := range strings.Fields(strings.ToLower(title)) {
		if len(word) > 3 && strings.Contains(firstHeading, word) {
			matches++
		}
	}
	return matches >= 2
}

func formattingTypeCount(s *PageSignals) int {
	present := []bool{
		s.ULCount > 0 || s.OLCount > 0,
		s.TableCount > 0,
		s.BlockquoteCount > 0,
		s.CodeCount > 0,
		s.DLCount > 0,
	}
	count := 0
	for _, p := range present {
		if p {
			count++
		}
	}
	return count
}
