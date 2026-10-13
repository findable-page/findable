package main

import (
	"fmt"
	"strings"
)

func evalSEO(s *PageSignals) []check {
	return runChecks(s,
		seoTitle,
		seoDescription,
		seoTechnicalMeta,
		seoHeadings,
		seoContent,
		seoLinks,
		seoImages,
		seoSocial,
		seoSecurityHeaders,
		seoStructure,
	)
}

func seoTitle(l *checkList, s *PageSignals) {
	if s.Title == "" {
		l.add("title_exists", "No title tag found", false)
		l.add("title_length", "Title is empty", false)
		return
	}
	l.add("title_exists", fmt.Sprintf("Title tag found: %q", s.Title), true)
	n := len(s.Title)
	if n >= 30 && n <= 60 {
		l.add("title_length", fmt.Sprintf("Title length: %d characters (optimal: 30-60)", n), true)
	} else {
		l.add("title_length", fmt.Sprintf("Title length: %d characters (aim for 30-60)", n), false)
	}
}

func seoDescription(l *checkList, s *PageSignals) {
	if s.MetaDesc == "" {
		l.add("desc_exists", "Missing meta description", false)
		l.add("desc_length", "Meta description is missing", false)
		return
	}
	n := len(s.MetaDesc)
	l.add("desc_exists", fmt.Sprintf("Meta description found (%d chars)", n), true)
	if n >= 120 && n <= 160 {
		l.add("desc_length", fmt.Sprintf("Description length: %d chars (optimal: 120-160)", n), true)
	} else {
		l.add("desc_length", fmt.Sprintf("Description length: %d chars (aim for 120-160)", n), false)
	}
}

func seoTechnicalMeta(l *checkList, s *PageSignals) {
	l.add("viewport", "Viewport meta tag found (mobile-friendly)", s.Viewport != "")
	l.add("charset", "UTF-8 charset declared", lowerContainsAny(s.Charset, "utf-8") || lowerContainsAny(s.RawHTML, "utf-8"))
	l.add("lang_attr", "Language attribute set on HTML tag", s.Lang != "")
	l.add("html_doctype", "<!DOCTYPE html> declared", s.HTMLDoctype)
	l.add("meta_refresh", "No meta refresh redirect (good)", s.MetaRefresh == "")
	l.add("robots_meta", "No robots meta tag blocking crawling", !lowerContainsAny(s.RobotsMeta, "noindex"))
}

func seoHeadings(l *checkList, s *PageSignals) {
	switch s.H1Count {
	case 1:
		l.add("h1_exists", "Single H1 tag found", true)
	case 0:
		l.add("h1_exists", "No H1 tag found", false)
	default:
		l.add("h1_exists", fmt.Sprintf("Multiple H1 tags found (%d) — exactly 1 recommended", s.H1Count), false)
	}
	l.add("h2_exists", fmt.Sprintf("%d H2 tags found", s.H2Count), s.H2Count > 0)
	l.add("heading_hierarchy", "Heading hierarchy is correct (no skipped levels)", headingsNeverSkipLevel(s.Headings))
}

func headingsNeverSkipLevel(headings []Heading) bool {
	for i := 1; i < len(headings); i++ {
		if headings[i].Level > headings[i-1].Level+1 {
			return false
		}
	}
	return true
}

func seoContent(l *checkList, s *PageSignals) {
	wordCount := len(s.Words)
	l.add("word_count", fmt.Sprintf("%d words in body (300+ recommended)", wordCount), wordCount >= 300)

	ratio := 0.0
	if s.RawHTML != "" {
		ratio = float64(len(s.TextContent)) / float64(len(s.RawHTML)) * 100
	}
	l.add("text_html_ratio", fmt.Sprintf("Text-to-HTML ratio: %.1f%% (>=15%% recommended)", ratio), ratio >= 15.0)
	l.add("dom_size", fmt.Sprintf("DOM size: ~%d elements (<1500 good)", s.DOMSize), s.DOMSize < 1500)
}

func seoLinks(l *checkList, s *PageSignals) {
	l.add("internal_links", fmt.Sprintf("%d internal links found", len(s.InternalLinks)), len(s.InternalLinks) >= 2)
	l.add("external_links", fmt.Sprintf("%d external links found", len(s.ExternalLinks)), len(s.ExternalLinks) >= 1)
	l.add("empty_hrefs", "No empty href attributes found", s.EmptyHrefs == 0)
	l.add("link_noopener", "External target=\"_blank\" links include rel=\"noopener\"", s.MissingOpenerCount == 0)
}

func seoImages(l *checkList, s *PageSignals) {
	if len(s.Images) == 0 {
		l.add("img_alt", "No images found on page", true)
		l.add("img_dimensions", "No images to check for dimensions", true)
		return
	}
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
	l.add("img_alt", fmt.Sprintf("%d of %d images missing alt attributes", missingAlt, len(s.Images)), missingAlt == 0)
	l.add("img_dimensions", fmt.Sprintf("%d of %d images missing width/height", missingDims, len(s.Images)), missingDims == 0)
}

func seoSocial(l *checkList, s *PageSignals) {
	l.add("canonical", "Canonical tag found", s.Canonical != "")
	l.add("og_title", "Open Graph title found", s.OGTags["og:title"] != "")
	l.add("og_desc", "Open Graph description found", s.OGTags["og:description"] != "")
	l.add("og_image", "Open Graph image found", s.OGTags["og:image"] != "")
	l.add("og_url", "Open Graph URL found", s.OGTags["og:url"] != "")
	l.add("twitter_card", "Twitter card declared", s.TwitterTags["twitter:card"] != "")
}

func seoSecurityHeaders(l *checkList, s *PageSignals) {
	l.add("https", "HTTPS enabled", s.URL != nil && s.URL.Scheme == "https")
	l.add("hsts", "Strict-Transport-Security header enforced", s.Headers.Get("Strict-Transport-Security") != "")
	l.add("x_content_type", "X-Content-Type-Options: nosniff header present", lowerContainsAny(s.Headers.Get("X-Content-Type-Options"), "nosniff"))
	l.add("x_frame_options", "X-Frame-Options header present", s.Headers.Get("X-Frame-Options") != "")
	l.add("csp", "Content-Security-Policy header declared", s.Headers.Get("Content-Security-Policy") != "")
}

func seoStructure(l *checkList, s *PageSignals) {
	l.add("jsonld", fmt.Sprintf("%d JSON-LD schema block(s) found", len(s.JSONLDs)), len(s.JSONLDs) > 0)
	l.add("aria_landmarks", "ARIA landmarks present", s.SemanticCounts["main"] > 0 && s.SemanticCounts["nav"] > 0)
	l.add("duplicate_ids", fmt.Sprintf("No duplicate IDs found (%d duplicates)", s.DuplicateIDs), s.DuplicateIDs == 0)
}
