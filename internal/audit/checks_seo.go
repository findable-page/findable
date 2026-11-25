package audit

import (
	"fmt"
	"slices"
	"strings"
)

func evalSEO(s *PageSignals) []Check {
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
		seoKeyphrase,
		seoHead,
		seoPage, seoSchema, seoSocialCompleteness,
	)
}

const (
	titleSelector = "head > title"
	tagImg        = "img"
	descSelector  = `head > meta[name=description]`
)

func seoTitle(l *checkList, s *PageSignals) {
	if s.Title == "" {
		l.addEv("title_exists", "No title tag found", false, evidence(titleSelector, ""))
		l.addEv("title_length", "Title is empty", false, evidence(titleSelector, ""))
		return
	}
	l.add("title_exists", fmt.Sprintf("Title tag found: %q", s.Title), true)
	n := len(s.Title)
	if n >= 30 && n <= 60 {
		l.add("title_length", fmt.Sprintf("Title length: %d characters (optimal: 30-60)", n), true)
	} else {
		l.addEv("title_length", fmt.Sprintf("Title length: %d characters (aim for 30-60)", n), false, evidence(titleSelector, s.Title))
	}
}

func seoDescription(l *checkList, s *PageSignals) {
	if s.MetaDesc == "" {
		l.addEv("desc_exists", "Missing meta description", false, evidence(descSelector, ""))
		l.addEv("desc_length", "Meta description is missing", false, evidence(descSelector, ""))
		return
	}
	n := len(s.MetaDesc)
	l.add("desc_exists", fmt.Sprintf("Meta description found (%d chars)", n), true)
	if n >= 120 && n <= 160 {
		l.add("desc_length", fmt.Sprintf("Description length: %d chars (optimal: 120-160)", n), true)
	} else {
		l.addEv("desc_length", fmt.Sprintf("Description length: %d chars (aim for 120-160)", n), false, evidence(descSelector, s.MetaDesc))
	}
}

func seoTechnicalMeta(l *checkList, s *PageSignals) {
	l.add("viewport", "Viewport meta tag found (mobile-friendly)", s.Viewport != "")
	l.add("charset", "UTF-8 charset declared", lowerContainsAny(s.Charset, "utf-8") || lowerContainsAny(s.RawHTML, "utf-8"))
	l.add("lang_attr", "Language attribute set on HTML tag", s.Lang != "")
	l.add("html_doctype", "<!DOCTYPE html> declared", s.HTMLDoctype)
	l.add("meta_refresh", "No meta refresh redirect (good)", s.MetaRefresh == "")
	l.addEv("robots_meta", "No robots meta tag blocking crawling", !lowerContainsAny(s.RobotsMeta, "noindex"), evidence(`head > meta[name=robots]`, s.RobotsMeta))
	xTagHeaders := s.Headers.Values("X-Robots-Tag")
	if len(xTagHeaders) == 0 {
		l.skip("x_robots_tag_noindex", "No X-Robots-Tag header to check")
	} else {
		xTag := strings.Join(xTagHeaders, ", ")
		l.addEv("x_robots_tag_noindex", "X-Robots-Tag header does not noindex this page", !ParseXRobotsTag(xTagHeaders).NoIndexFor(pageCrawler), evidence("X-Robots-Tag", xTag))
	}
}

func seoHeadings(l *checkList, s *PageSignals) {
	switch s.H1Count {
	case 1:
		l.add("h1_exists", "Single H1 tag found", true)
	case 0:
		l.addEv("h1_exists", "No H1 tag found", false, evidence("body h1", ""))
	default:
		evs := make([]Evidence, 0, len(s.Tags["h1"]))
		for _, t := range s.Tags["h1"] {
			evs = append(evs, evidence(t.Selector, t.Text))
		}
		l.addEv("h1_exists", fmt.Sprintf("Multiple H1 tags found (%d) — exactly 1 recommended", s.H1Count), false, evs...)
	}
	l.add("h2_exists", fmt.Sprintf("%d H2 tags found", s.H2Count), s.H2Count > 0)
	l.addEv("heading_hierarchy", "Heading hierarchy is correct (no skipped levels)", headingsNeverSkipLevel(s.Headings), firstSkipEvidence(s)...)
}

func firstSkipEvidence(s *PageSignals) []Evidence {
	tags := headingTags(s)
	for i := 1; i < len(s.Headings) && i < len(tags); i++ {
		if s.Headings[i].Level > s.Headings[i-1].Level+1 {
			return []Evidence{evidence(tags[i].Selector, s.Headings[i].Text)}
		}
	}
	return nil
}

// headingTags returns h1-h6 TagInfos in document order (Pos order), matching
// the order of s.Headings.
func headingTags(s *PageSignals) []TagInfo {
	var out []TagInfo
	for level := 1; level <= 6; level++ {
		out = append(out, s.Tags[fmt.Sprintf("h%d", level)]...)
	}
	slices.SortStableFunc(out, func(a, b TagInfo) int { return a.Pos - b.Pos })
	return out
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
	l.addEv("empty_hrefs", "No empty href attributes found", s.EmptyHrefs == 0, emptyHrefEvidence(s)...)
	l.addEv("link_noopener", "External target=\"_blank\" links include rel=\"noopener\"", s.MissingOpenerCount == 0, openerEvidence(s)...)
}

func emptyHrefEvidence(s *PageSignals) []Evidence {
	var out []Evidence
	for _, a := range s.Tags["a"] {
		href := strings.TrimSpace(a.Attrs["href"])
		if href == "" || href == "#" {
			out = append(out, evidence(a.Selector, href))
		}
	}
	return out
}

func openerEvidence(s *PageSignals) []Evidence {
	var out []Evidence
	for _, a := range s.Tags["a"] {
		rel := strings.ToLower(a.Attrs["rel"])
		if strings.EqualFold(strings.TrimSpace(a.Attrs["target"]), "_blank") && !strings.Contains(rel, "noopener") && !strings.Contains(rel, "noreferrer") {
			out = append(out, evidence(a.Selector, a.Attrs["href"]))
		}
	}
	return out
}

func seoImages(l *checkList, s *PageSignals) {
	if len(s.Images) == 0 {
		l.add("img_alt", "No images found on page", true)
		l.add("img_dimensions", "No images to check for dimensions", true)
		return
	}
	imgTags := s.Tags[tagImg]
	imgSelector := func(i int) string {
		if i < len(imgTags) {
			return imgTags[i].Selector
		}
		return tagImg
	}
	var missingAlt, missingDims []Evidence
	for i, img := range s.Images {
		if strings.TrimSpace(img.Alt) == "" {
			missingAlt = append(missingAlt, evidence(imgSelector(i), img.Src))
		}
		if img.Width == "" || img.Height == "" {
			missingDims = append(missingDims, evidence(imgSelector(i), img.Src))
		}
	}
	l.addEv("img_alt", fmt.Sprintf("%d of %d images missing alt attributes", len(missingAlt), len(s.Images)), len(missingAlt) == 0, missingAlt...)
	l.addEv("img_dimensions", fmt.Sprintf("%d of %d images missing width/height", len(missingDims), len(s.Images)), len(missingDims) == 0, missingDims...)
}

func seoSocial(l *checkList, s *PageSignals) {
	l.addEv("canonical", "Canonical tag found", s.Canonical != "", evidence(`head > link[rel=canonical]`, s.Canonical))
	og := func(id, label, property string) {
		l.addEv(id, label, s.OGTags[property] != "", evidence("head > meta[property="+property+"]", s.OGTags[property]))
	}
	og("og_title", "Open Graph title found", "og:title")
	og("og_desc", "Open Graph description found", "og:description")
	og("og_image", "Open Graph image found", "og:image")
	og("og_url", "Open Graph URL found", "og:url")
	l.addEv("twitter_card", "Twitter card declared", s.TwitterTags["twitter:card"] != "", evidence(`head > meta[name=twitter:card]`, s.TwitterTags["twitter:card"]))
}

func seoSecurityHeaders(l *checkList, s *PageSignals) {
	l.add("https", "HTTPS enabled", s.URL != nil && s.URL.Scheme == schemeHTTPS)
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

func seoKeyphrase(l *checkList, s *PageSignals) {
	kp := strings.ToLower(s.Keyphrase)
	if s.Keyphrase == "" {
		return
	}
	l.add("keyphrase_in_title", `Focus keyphrase "`+s.Keyphrase+`" in title tag`, lowerContainsAny(s.Title, kp))
	l.add("keyphrase_in_desc", `Focus keyphrase "`+s.Keyphrase+`" in meta description`, lowerContainsAny(s.MetaDesc, kp))
	l.add("keyphrase_in_h1", `Focus keyphrase "`+s.Keyphrase+`" in H1 heading`, h1Contains(s.Headings, kp))
	l.add("keyphrase_in_content", `Focus keyphrase "`+s.Keyphrase+`" appears in body copy`, lowerContainsAny(s.TextContent, kp))
}

func h1Contains(headings []Heading, keyphrase string) bool {
	for _, h := range headings {
		if h.Level == 1 && lowerContainsAny(h.Text, keyphrase) {
			return true
		}
	}
	return false
}
