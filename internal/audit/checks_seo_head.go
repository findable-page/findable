package audit

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	langTagRegex    = regexp.MustCompile(`^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)
	hreflangRegex   = regexp.MustCompile(`(?i)^(x-default|[a-z]{2,3}(-[a-z]{4})?(-([a-z]{2}|\d{3}))?)$`)
	maxScaleRegex   = regexp.MustCompile(`maximum-scale=(\d+(?:\.\d+)?)`)
	charsetTagRegex = regexp.MustCompile(`(?i)<meta[^>]+charset|http-equiv=["']?content-type`)
)

func seoHead(l *checkList, s *PageSignals) {
	seoCanonicalQuality(l, s)
	seoRobotsDirectives(l, s)
	seoSingletons(l, s)
	seoURLShape(l, s)
	seoBrandAssets(l, s)
	seoViewportAndLang(l, s)
	seoHreflang(l, s)
	seoCharsetPlacement(l, s)
}

func linkTags(s *PageSignals, rel string) []TagInfo {
	var out []TagInfo
	for _, t := range s.Tags["link"] {
		for _, token := range strings.Fields(strings.ToLower(t.Attrs["rel"])) {
			if token == rel {
				out = append(out, t)
				break
			}
		}
	}
	return out
}

func metaTags(s *PageSignals, name string) []string {
	var out []string
	for _, t := range s.Tags["meta"] {
		if strings.ToLower(strings.TrimSpace(t.Attrs["name"])) == name {
			out = append(out, strings.TrimSpace(t.Attrs["content"]))
		}
	}
	return out
}

func firstHref(tags []TagInfo) string {
	for _, t := range tags {
		if href := strings.TrimSpace(t.Attrs["href"]); href != "" {
			return href
		}
	}
	return ""
}

func pagePathKey(u *url.URL) string {
	return strings.ToLower(u.Host) + strings.TrimSuffix(u.Path, "/")
}

func seoCanonicalQuality(l *checkList, s *PageSignals) {
	canonical := firstHref(linkTags(s, "canonical"))
	if canonical == "" {
		l.add("canonical_self", "No canonical to compare with this page", true)
		l.add("canonical_absolute", "No canonical to validate", true)
		return
	}
	target, err := url.Parse(canonical)
	self := err == nil && s.URL != nil && pagePathKey(s.URL.ResolveReference(target)) == pagePathKey(s.URL)
	if self {
		l.add("canonical_self", "Canonical points to this page", true)
	} else {
		l.add("canonical_self", fmt.Sprintf("Canonical points elsewhere: %s", canonical), false)
	}
	l.add("canonical_absolute", "Canonical is an absolute URL", err == nil && target.IsAbs())
}

func robotsTokens(s *PageSignals) map[string]bool {
	tokens := map[string]bool{}
	for _, content := range metaTags(s, "robots") {
		for _, token := range strings.Split(strings.ToLower(content), ",") {
			tokens[strings.TrimSpace(token)] = true
		}
	}
	return tokens
}

func seoRobotsDirectives(l *checkList, s *PageSignals) {
	tokens := robotsTokens(s)
	conflict := (tokens["index"] && tokens["noindex"]) || (tokens["follow"] && tokens["nofollow"])
	l.add("meta_robots_conflict", "Robots meta directives do not contradict each other", !conflict)
	l.add("robots_nofollow", "Page does not nofollow all its links via robots meta", !tokens["nofollow"] && !tokens["none"])
}

func seoSingletons(l *checkList, s *PageSignals) {
	titles := 0
	for _, t := range s.Tags["title"] {
		if t.InHead {
			titles++
		}
	}
	l.add("title_single", fmt.Sprintf("%d title tag(s) in head (exactly 1 expected)", titles), titles <= 1)
	descriptions := len(metaTags(s, "description"))
	l.add("desc_single", fmt.Sprintf("%d meta description tag(s) (exactly 1 expected)", descriptions), descriptions <= 1)
}

func seoURLShape(l *checkList, s *PageSignals) {
	if s.URL == nil {
		return
	}
	requestURI := s.URL.RequestURI()
	l.add("url_length", fmt.Sprintf("URL path and query are %d characters (100 or fewer recommended)", len(requestURI)), len(requestURI) <= 100)

	params := 0
	for _, pair := range strings.Split(s.URL.RawQuery, "&") {
		if pair != "" {
			params++
		}
	}
	l.add("url_query_params", fmt.Sprintf("%d query parameter(s) in URL (2 or fewer recommended)", params), params <= 2)

	depth := 0
	for _, segment := range strings.Split(s.URL.Path, "/") {
		if segment != "" {
			depth++
		}
	}
	l.add("url_depth", fmt.Sprintf("URL is %d level(s) deep (5 or fewer recommended)", depth), depth <= 5)
}

func seoBrandAssets(l *checkList, s *PageSignals) {
	l.add("favicon", "Favicon link found", firstHref(linkTags(s, "icon")) != "")
	l.add("apple_touch_icon", "Apple touch icon declared", firstHref(linkTags(s, "apple-touch-icon")) != "")
	themeColors := metaTags(s, "theme-color")
	l.add("theme_color", "theme-color meta tag found", len(themeColors) > 0 && themeColors[0] != "")
	l.add("web_manifest", "Web app manifest linked", firstHref(linkTags(s, "manifest")) != "")
	feedType := false
	for _, t := range linkTags(s, "alternate") {
		feedType = feedType || lowerContainsAny(t.Attrs["type"], "rss", "atom")
	}
	l.add("rss_feed", "RSS or Atom feed advertised", feedType)
}

func seoViewportAndLang(l *checkList, s *PageSignals) {
	if viewport := strings.ToLower(strings.ReplaceAll(s.Viewport, " ", "")); viewport == "" {
		l.add("viewport_device_width", "No viewport to validate", true)
		l.add("viewport_scalable", "No viewport to validate", true)
	} else {
		l.add("viewport_device_width", "Viewport sets width=device-width", strings.Contains(viewport, "width=device-width"))
		l.add("viewport_scalable", "Viewport does not block pinch zoom", !blocksZoom(viewport))
	}
	if s.Lang == "" {
		l.add("lang_valid", "No lang attribute to validate", true)
		return
	}
	l.add("lang_valid", fmt.Sprintf("HTML lang %q is a valid language tag", s.Lang), langTagRegex.MatchString(s.Lang))
}

func blocksZoom(viewport string) bool {
	if strings.Contains(viewport, "user-scalable=no") || strings.Contains(viewport, "user-scalable=0") {
		return true
	}
	if m := maxScaleRegex.FindStringSubmatch(viewport); m != nil {
		scale, _ := strconv.ParseFloat(m[1], 64)
		return scale < 2
	}
	return false
}

func seoHreflang(l *checkList, s *PageSignals) {
	invalid, relative, total := 0, 0, 0
	for _, t := range linkTags(s, "alternate") {
		code := strings.TrimSpace(t.Attrs["hreflang"])
		if code == "" {
			continue
		}
		total++
		if !hreflangRegex.MatchString(code) {
			invalid++
		}
		if href := strings.ToLower(strings.TrimSpace(t.Attrs["href"])); !strings.HasPrefix(href, "http://") && !strings.HasPrefix(href, "https://") {
			relative++
		}
	}
	if total == 0 {
		l.add("hreflang_valid", "No hreflang annotations to validate", true)
		l.add("hreflang_absolute", "No hreflang annotations to validate", true)
		return
	}
	l.add("hreflang_valid", fmt.Sprintf("%d of %d hreflang codes are invalid", invalid, total), invalid == 0)
	l.add("hreflang_absolute", fmt.Sprintf("%d of %d hreflang URLs are not absolute", relative, total), relative == 0)
}

func seoCharsetPlacement(l *checkList, s *PageSignals) {
	if !charsetTagRegex.MatchString(s.RawHTML) {
		l.add("charset_early", "No charset declaration to place", true)
		return
	}
	head := s.RawHTML
	if len(head) > 1024 {
		head = head[:1024]
	}
	l.add("charset_early", "Charset declared within the first 1024 bytes", charsetTagRegex.MatchString(head))
}
