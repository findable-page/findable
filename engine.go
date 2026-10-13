package main

import (
	"fmt"
	"math"
	"net/http"
	"strings"
)

func AuditPage(targetURL string, rawHTML string, headers http.Header, statusCode int) scan {
	sig := extractSignals(targetURL, rawHTML, headers, statusCode)

	seoChecks := evalSEO(sig)
	aeoChecks := evalAEO(sig)
	geoChecks := evalGEO(sig)

	s := scan{
		SEO:           buildCategory(seoChecks),
		AEO:           buildCategory(aeoChecks),
		GEO:           buildCategory(geoChecks),
		CWV:           category{Checks: []check{}},
		InternalLinks: sig.InternalLinks,
	}

	return s
}

func buildCategory(checks []check) category {
	passed := 0
	total := 0
	for _, c := range checks {
		if c.Applicable != nil && !*c.Applicable {
			continue
		}
		total++
		if c.Pass {
			passed++
		}
	}
	var scoreVal float64
	if total > 0 {
		scoreVal = math.Round((float64(passed) / float64(total)) * 100)
	}
	return category{
		Score:  &scoreVal,
		Passed: passed,
		Total:  total,
		Checks: checks,
	}
}

func evalSEO(s PageSignals) []check {
	var checks []check
	chk := func(id, label string, pass bool, applicable *bool) {
		checks = append(checks, check{ID: id, Label: label, Pass: pass, Applicable: applicable})
	}

	// 1. Title
	if s.Title == "" {
		chk("title_exists", "No title tag found", false, nil)
		chk("title_length", "Title is empty", false, nil)
	} else {
		chk("title_exists", fmt.Sprintf("Title tag found: %q", s.Title), true, nil)
		tLen := len(s.Title)
		if tLen >= 30 && tLen <= 60 {
			chk("title_length", fmt.Sprintf("Title length: %d characters (optimal: 30-60)", tLen), true, nil)
		} else {
			chk("title_length", fmt.Sprintf("Title length: %d characters (aim for 30-60)", tLen), false, nil)
		}
	}

	// 2. Meta description
	if s.MetaDesc == "" {
		chk("desc_exists", "Missing meta description", false, nil)
		chk("desc_length", "Meta description is missing", false, nil)
	} else {
		chk("desc_exists", fmt.Sprintf("Meta description found (%d chars)", len(s.MetaDesc)), true, nil)
		dLen := len(s.MetaDesc)
		if dLen >= 120 && dLen <= 160 {
			chk("desc_length", fmt.Sprintf("Description length: %d chars (optimal: 120-160)", dLen), true, nil)
		} else {
			chk("desc_length", fmt.Sprintf("Description length: %d chars (aim for 120-160)", dLen), false, nil)
		}
	}

	// 3. Technical meta & HTML
	chk("viewport", "Viewport meta tag found (mobile-friendly)", s.Viewport != "", nil)
	chk("charset", "UTF-8 charset declared", strings.Contains(strings.ToLower(s.Charset), "utf-8") || strings.Contains(strings.ToLower(s.RawHTML), "utf-8"), nil)
	chk("lang_attr", "Language attribute set on HTML tag", s.Lang != "", nil)
	chk("html_doctype", "<!DOCTYPE html> declared", s.HTMLDoctype, nil)
	chk("meta_refresh", "No meta refresh redirect (good)", s.MetaRefresh == "", nil)
	chk("robots_meta", "No robots meta tag blocking crawling", !strings.Contains(strings.ToLower(s.RobotsMeta), "noindex"), nil)

	// 4. Headings & Content
	if s.H1Count == 1 {
		chk("h1_exists", "Single H1 tag found", true, nil)
	} else if s.H1Count == 0 {
		chk("h1_exists", "No H1 tag found", false, nil)
	} else {
		chk("h1_exists", fmt.Sprintf("Multiple H1 tags found (%d) — exactly 1 recommended", s.H1Count), false, nil)
	}
	chk("h2_exists", fmt.Sprintf("%d H2 tags found", s.H2Count), s.H2Count > 0, nil)

	// Heading hierarchy
	hierarchyOK := true
	for i := 1; i < len(s.Headings); i++ {
		if s.Headings[i].Level > s.Headings[i-1].Level+1 {
			hierarchyOK = false
			break
		}
	}
	chk("heading_hierarchy", "Heading hierarchy is correct (no skipped levels)", hierarchyOK, nil)

	// Word count & text/HTML ratio
	wordCount := len(s.Words)
	chk("word_count", fmt.Sprintf("%d words in body (300+ recommended)", wordCount), wordCount >= 300, nil)

	ratio := 0.0
	if len(s.RawHTML) > 0 {
		ratio = float64(len(s.TextContent)) / float64(len(s.RawHTML)) * 100
	}
	chk("text_html_ratio", fmt.Sprintf("Text-to-HTML ratio: %.1f%% (>=15%% recommended)", ratio), ratio >= 15.0, nil)
	chk("dom_size", fmt.Sprintf("DOM size: ~%d elements (<1500 good)", s.DOMSize), s.DOMSize < 1500, nil)

	// 5. Links
	chk("internal_links", fmt.Sprintf("%d internal links found", len(s.InternalLinks)), len(s.InternalLinks) >= 2, nil)
	chk("external_links", fmt.Sprintf("%d external links found", len(s.ExternalLinks)), len(s.ExternalLinks) >= 1, nil)
	chk("empty_hrefs", "No empty href attributes found", s.EmptyHrefs == 0, nil)
	chk("link_noopener", "External target=\"_blank\" links include rel=\"noopener\"", s.MissingOpenerCount == 0, nil)

	// 6. Images
	if len(s.Images) == 0 {
		chk("img_alt", "No images found on page", true, nil)
		chk("img_dimensions", "No images to check for dimensions", true, nil)
	} else {
		missingAlt := 0
		missingDims := 0
		for _, img := range s.Images {
			if strings.TrimSpace(img.Alt) == "" {
				missingAlt++
			}
			if img.Width == "" || img.Height == "" {
				missingDims++
			}
		}
		chk("img_alt", fmt.Sprintf("%d of %d images missing alt attributes", missingAlt, len(s.Images)), missingAlt == 0, nil)
		chk("img_dimensions", fmt.Sprintf("%d of %d images missing width/height", missingDims, len(s.Images)), missingDims == 0, nil)
	}

	// 7. Canonical & Social
	chk("canonical", "Canonical tag found", s.Canonical != "", nil)
	chk("og_title", "Open Graph title found", s.OGTags["og:title"] != "", nil)
	chk("og_desc", "Open Graph description found", s.OGTags["og:description"] != "", nil)
	chk("og_image", "Open Graph image found", s.OGTags["og:image"] != "", nil)
	chk("og_url", "Open Graph URL found", s.OGTags["og:url"] != "", nil)
	chk("twitter_card", "Twitter card declared", s.TwitterTags["twitter:card"] != "", nil)

	// 8. Technical Security & Headers
	isHTTPS := s.URL != nil && s.URL.Scheme == "https"
	chk("https", "HTTPS enabled", isHTTPS, nil)
	chk("hsts", "Strict-Transport-Security header enforced", s.Headers.Get("Strict-Transport-Security") != "", nil)
	chk("x_content_type", "X-Content-Type-Options: nosniff header present", strings.Contains(strings.ToLower(s.Headers.Get("X-Content-Type-Options")), "nosniff"), nil)
	chk("x_frame_options", "X-Frame-Options header present", s.Headers.Get("X-Frame-Options") != "", nil)
	chk("csp", "Content-Security-Policy header declared", s.Headers.Get("Content-Security-Policy") != "", nil)

	// 9. Structured Data & Accessibility
	chk("jsonld", fmt.Sprintf("%d JSON-LD schema block(s) found", len(s.JSONLDs)), len(s.JSONLDs) > 0, nil)
	chk("aria_landmarks", "ARIA landmarks present", s.SemanticCounts["main"] > 0 && s.SemanticCounts["nav"] > 0, nil)
	chk("duplicate_ids", fmt.Sprintf("No duplicate IDs found (%d duplicates)", s.DuplicateIDs), s.DuplicateIDs == 0, nil)

	return checks
}

func evalAEO(s PageSignals) []check {
	var checks []check
	chk := func(id, label string, pass bool, applicable *bool) {
		checks = append(checks, check{ID: id, Label: label, Pass: pass, Applicable: applicable})
	}

	// 1. Structured data & FAQ
	hasFAQ := false
	for _, st := range s.SchemaTypes {
		if strings.EqualFold(st, "FAQPage") {
			hasFAQ = true
			break
		}
	}
	chk("aeo_structured_data", "Structured data found for AI extraction", len(s.JSONLDs) > 0, nil)
	chk("aeo_faq", "FAQ schema or Q&A section detected", hasFAQ, nil)

	// 2. Direct answer paragraphs under headings
	hasDirectAnswer := false
	for _, p := range s.Paragraphs {
		pWords := len(strings.Fields(p))
		if pWords >= 30 && pWords <= 80 {
			hasDirectAnswer = true
			break
		}
	}
	chk("aeo_direct_answers", "Concise answer paragraphs (30-80 words) found", hasDirectAnswer, nil)

	// 3. Question-style headings
	hasQHeadings := false
	for _, h := range s.Headings {
		t := strings.ToLower(h.Text)
		if strings.HasPrefix(t, "what") || strings.HasPrefix(t, "how") || strings.HasPrefix(t, "why") || strings.HasSuffix(t, "?") {
			hasQHeadings = true
			break
		}
	}
	chk("aeo_question_headings", "Question-style headings found (What/How/Why/?)", hasQHeadings, nil)

	// 4. Structural formatting
	chk("aeo_definition_lists", "Definition lists (<dl>) present for terminology", s.DLCount > 0, nil)
	chk("aeo_numbered_steps", "Numbered steps (<ol>) present for instructions", s.OLCount > 0, nil)
	chk("aeo_tables", "Data tables present for comparison", s.TableCount > 0, nil)
	chk("aeo_lists", "Lists (<ul>/<ol>) present for scannability", s.ULCount > 0 || s.OLCount > 0, nil)
	chk("aeo_code_blocks", "Code blocks (<code>/<pre>) present for technical clarity", s.CodeCount > 0, nil)
	chk("aeo_collapsible", "<details>/<summary> elements present", s.DetailsCount > 0, nil)

	// 5. Readability & Tone
	lowerText := strings.ToLower(s.TextContent)
	youCount := strings.Count(lowerText, " you ") + strings.Count(lowerText, " your ") + strings.Count(lowerText, " we ")
	chk("aeo_conversational", "Conversational tone with user pronouns ('you', 'your', 'we')", youCount >= 3, nil)

	sentences := strings.Split(s.TextContent, ".")
	avgSentenceLen := 15.0
	if len(sentences) > 0 {
		avgSentenceLen = float64(len(s.Words)) / float64(len(sentences))
	}
	chk("aeo_reading_level", fmt.Sprintf("Average sentence length: %.1f words (<20 is AI-friendly)", avgSentenceLen), avgSentenceLen <= 22.0, nil)
	chk("aeo_short_sentences", "Concise sentence structures dominant", avgSentenceLen <= 20.0, nil)

	hasSummary := strings.Contains(lowerText, "key takeaway") || strings.Contains(lowerText, "summary") || strings.Contains(lowerText, "conclusion")
	chk("aeo_summary", "Summary or Key Takeaways section present", hasSummary, nil)

	// 6. Clean URL & Semantic HTML
	cleanURL := s.URL != nil && !strings.Contains(s.URL.Path, "_") && strings.ToLower(s.URL.Path) == s.URL.Path
	chk("aeo_clean_url", "Clean lowercase URL structure", cleanURL, nil)

	semanticTotal := 0
	for _, count := range s.SemanticCounts {
		semanticTotal += count
	}
	chk("aeo_semantic_html", fmt.Sprintf("%d semantic HTML5 elements found", semanticTotal), semanticTotal >= 4, nil)

	// 7. Robots & AI
	chk("aeo_ai_robots", "No AI-blocking directives detected", !strings.Contains(strings.ToLower(s.RobotsMeta), "noindex"), nil)

	return checks
}

func evalGEO(s PageSignals) []check {
	var checks []check
	chk := func(id, label string, pass bool, applicable *bool) {
		checks = append(checks, check{ID: id, Label: label, Pass: pass, Applicable: applicable})
	}

	// 1. Empirical statistics & data
	numStats := len(s.NumbersWithUnits)
	chk("geo_statistics_density", fmt.Sprintf("%d quantitative data points / metrics found", numStats), numStats >= 3, nil)
	chk("geo_data_viz", "Structured data tables or charts found", s.TableCount > 0, nil)
	chk("geo_comparison", "Comparative sections or tables detected", s.TableCount > 0 || strings.Contains(strings.ToLower(s.TextContent), " vs "), nil)

	// 2. Expert quotes & citations
	chk("geo_expert_quotes", fmt.Sprintf("%d expert quotes/blockquotes found", len(s.ExpertQuotes)), len(s.ExpertQuotes) > 0, nil)
	chk("geo_citations", fmt.Sprintf("%d external source citations found", len(s.ExternalLinks)), len(s.ExternalLinks) > 0, nil)

	// 3. Definitions & Passage Citability
	hasDef := strings.Contains(strings.ToLower(s.TextContent), " is a ") || strings.Contains(strings.ToLower(s.TextContent), " is the ") || strings.Contains(strings.ToLower(s.TextContent), " defined as ")
	chk("geo_definitions", "Clear concept definitions found ('X is a...')", hasDef, nil)

	hasPassage := false
	for _, p := range s.Paragraphs {
		w := len(strings.Fields(p))
		if w >= 40 && w <= 120 {
			hasPassage = true
			break
		}
	}
	chk("geo_passage_citability", "Passage-ready standalone answer blocks present", hasPassage, nil)

	// 4. Authority & Schema Graph
	hasPersonAuthor := false
	hasOrg := false
	for _, st := range s.SchemaTypes {
		if strings.EqualFold(st, "Person") {
			hasPersonAuthor = true
		}
		if strings.EqualFold(st, "Organization") || strings.EqualFold(st, "Bakery") || strings.EqualFold(st, "LocalBusiness") {
			hasOrg = true
		}
	}
	chk("geo_author_schema", "Author schema markup found", hasPersonAuthor || s.AuthorInSchema != "", nil)
	chk("geo_author_credentials", "Visible author byline with credentials or dates", strings.Contains(strings.ToLower(s.TextContent), "by ") || s.AuthorInSchema != "", nil)
	chk("geo_org_schema", "Organization or Business schema markup declared", hasOrg, nil)
	chk("geo_knowledge_graph", fmt.Sprintf("%d authoritative sameAs links in schema", len(s.SameAsLinks)), len(s.SameAsLinks) > 0, nil)

	// Trust pages
	hasAbout := false
	hasContact := false
	for _, l := range s.InternalLinks {
		lowerL := strings.ToLower(l)
		if strings.Contains(lowerL, "about") {
			hasAbout = true
		}
		if strings.Contains(lowerL, "contact") {
			hasContact = true
		}
	}
	chk("geo_trust", "Trust signals (About and Contact links) present", hasAbout && hasContact, nil)
	chk("geo_entity_markup", fmt.Sprintf("%d schema entity type(s) declared", len(s.SchemaTypes)), len(s.SchemaTypes) >= 2, nil)

	// 5. Freshness & Content Depth
	hasFreshness := s.SemanticCounts["time"] > 0 || yearRegex.MatchString(s.TextContent) || s.DateModified != ""
	chk("geo_freshness", "Content freshness signals (dates/time elements) present", hasFreshness, nil)
	chk("geo_time_element", "<time> machine-readable tags present", s.SemanticCounts["time"] > 0, nil)
	chk("geo_content_length", fmt.Sprintf("%d words content depth (300+ recommended)", len(s.Words)), len(s.Words) >= 300, nil)

	// Intent clarity
	hasIntentClarity := false
	if s.Title != "" && len(s.Headings) > 0 {
		titleWords := strings.Fields(strings.ToLower(s.Title))
		h1Lower := strings.ToLower(s.Headings[0].Text)
		matches := 0
		for _, tw := range titleWords {
			if len(tw) > 3 && strings.Contains(h1Lower, tw) {
				matches++
			}
		}
		hasIntentClarity = matches >= 2
	}
	chk("geo_intent_clarity", "Title keywords match primary H1 heading", hasIntentClarity, nil)

	// Formatting variety
	formatTypes := 0
	if s.ULCount > 0 || s.OLCount > 0 {
		formatTypes++
	}
	if s.TableCount > 0 {
		formatTypes++
	}
	if s.BlockquoteCount > 0 {
		formatTypes++
	}
	if s.CodeCount > 0 {
		formatTypes++
	}
	if s.DLCount > 0 {
		formatTypes++
	}
	chk("geo_format_variety", fmt.Sprintf("%d formatting types used (tables, lists, quotes, code, dl)", formatTypes), formatTypes >= 3, nil)
	chk("geo_snippet_control", "No restrictive nosnippet over-blocking", !strings.Contains(strings.ToLower(s.RobotsMeta), "nosnippet"), nil)

	return checks
}
