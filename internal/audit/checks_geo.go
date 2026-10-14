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
	l.add("geo_intent_clarity", "Title keywords match primary H1 heading", titleMatchesFirstHeading(s))
	formats := formattingTypeCount(s)
	l.add("geo_format_variety", fmt.Sprintf("%d formatting types used (tables, lists, quotes, code, dl)", formats), formats >= 3)
	l.add("geo_snippet_control", "No restrictive nosnippet over-blocking", !lowerContainsAny(s.RobotsMeta, "nosnippet"))
}

func titleMatchesFirstHeading(s *PageSignals) bool {
	if s.Title == "" || len(s.Headings) == 0 {
		return false
	}
	firstHeading := strings.ToLower(s.Headings[0].Text)
	matches := 0
	for _, word := range strings.Fields(strings.ToLower(s.Title)) {
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
