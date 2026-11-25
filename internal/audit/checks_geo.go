package audit

import (
	"fmt"
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
	l.add("geo_knowledge_graph", fmt.Sprintf("%d authoritative sameAs links in schema", len(s.SameAsLinks)), len(s.SameAsLinks) > 0)
	geoSchemaKnowsAbout(l, s)
	geoAuthorCredentialDepth(l, s)
}

const knowsAboutMinimum = 3

var credentialKeys = []string{"jobTitle", "alumniOf", "hasCredential"}

// professionalProfileHints are sameAs hosts that signal an academic or
// professional identity rather than a generic social profile.
var professionalProfileHints = []string{"orcid.org", "scholar.google.com", "researchgate.net", "academia.edu", "linkedin.com"}

// geoSchemaKnowsAbout rewards topical authority signaling: schema.org
// documents knowsAbout on Person and Organization, and GEO guides also place
// it on WebSite nodes, so both are accepted here.
func geoSchemaKnowsAbout(l *checkList, s *PageSignals) {
	doc := parseSchemaDoc(s)
	nodes := doc.ofType(append(append([]string{}, orgTypes...), localTypes...)...)
	nodes = append(nodes, doc.ofType("website")...)
	if len(nodes) == 0 {
		l.skip("geo_schema_knows_about", "No Organization or WebSite schema to validate")
		return
	}
	seen := map[string]bool{}
	for _, node := range nodes {
		for _, entry := range jsonItems(node["knowsAbout"]) {
			if topic := schemaTextOrName(entry); topic != "" && !seen[strings.ToLower(topic)] {
				seen[strings.ToLower(topic)] = true
			}
		}
	}
	l.add("geo_schema_knows_about", fmt.Sprintf("%d distinct topic(s) in knowsAbout on Organization or WebSite schema (%d or more recommended)", len(seen), knowsAboutMinimum), len(seen) >= knowsAboutMinimum)
}

func schemaTextOrName(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case map[string]any:
		if name, ok := t["name"].(string); ok {
			return strings.TrimSpace(name)
		}
	}
	return ""
}

// geoAuthorCredentialDepth checks that author Person entities in schema carry
// a verifiable credential: jobTitle, alumniOf, hasCredential, or an academic
// or professional profile in sameAs.
func geoAuthorCredentialDepth(l *checkList, s *PageSignals) {
	authors := schemaAuthorEntities(parseSchemaDoc(s))
	if len(authors) == 0 {
		l.skip("geo_author_credential_depth", "No author Person schema to validate")
		return
	}
	credentialed := 0
	for _, author := range authors {
		if authorHasCredential(author) {
			credentialed++
		}
	}
	l.add("geo_author_credential_depth", fmt.Sprintf("%d of %d schema author(s) carry a credential (jobTitle, alumniOf, hasCredential or a professional profile)", credentialed, len(authors)), credentialed == len(authors))
}

// schemaAuthorEntities collects Person authors from every schema node: map
// values (typed Person or untyped, whose default type is Person), and @id or
// URL references resolved through the document's nodes. Explicitly
// other-typed authors (such as Organization bylines) are out of scope.
func schemaAuthorEntities(doc schemaDoc) []map[string]any {
	var authors []map[string]any
	for _, node := range doc.nodes {
		for _, value := range jsonItems(node["author"]) {
			switch a := value.(type) {
			case map[string]any:
				if id, _ := a["@id"].(string); id != "" && !jsonPresent(a["name"]) && len(nodeTypes(a)) == 0 {
					if target := doc.nodeByID(id); target != nil && authorEntityIsPerson(target) {
						authors = append(authors, target)
					}
				} else if authorEntityIsPerson(a) {
					authors = append(authors, a)
				}
			case string:
				if target := doc.nodeByID(strings.TrimSpace(a)); target != nil && authorEntityIsPerson(target) {
					authors = append(authors, target)
				}
			}
		}
	}
	return authors
}

// authorEntityIsPerson accepts Person-typed entities and untyped author
// objects, whose default type is Person. Explicitly other-typed authors
// (such as Organization bylines) are out of scope.
func authorEntityIsPerson(m map[string]any) bool {
	types := nodeTypes(m)
	if len(types) == 0 {
		return true
	}
	return containsString(types, "person")
}

func (d schemaDoc) nodeByID(id string) map[string]any {
	for _, node := range d.nodes {
		if value, _ := node["@id"].(string); value == id {
			return node
		}
	}
	return nil
}

func authorHasCredential(author map[string]any) bool {
	for _, key := range credentialKeys {
		if jsonPresent(author[key]) {
			return true
		}
	}
	for _, link := range sameAsLinks(author["sameAs"]) {
		lower := strings.ToLower(link)
		for _, hint := range professionalProfileHints {
			if strings.Contains(lower, hint) {
				return true
			}
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
