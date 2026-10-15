package audit

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	altFilenameRegex    = regexp.MustCompile(`(?i)\.(jpe?g|png|gif|webp|avif|svg|bmp)$|^(image|img|photo|picture|pic|untitled|screenshot)[-_ ]?\d*$|^(img|dsc|dscn|screenshot)[-_]\d+`)
	imageExtRegex       = regexp.MustCompile(`(?i)\.(jpe?g|png|gif|webp|avif|svg)$`)
	modernImageExtRegex = regexp.MustCompile(`(?i)\.(webp|avif|svg)$`)
	genericAnchors      = map[string]bool{"click here": true, "here": true, "read more": true, "more": true, "learn more": true, "link": true, "this": true, "this link": true, "click": true}
)

const maxEagerImages = 5

func seoPage(l *checkList, s *PageSignals) {
	seoImageQuality(l, s)
	seoLinkQuality(l, s)
	seoLoadPerformance(l, s)
	seoAccessibility(l, s)
}

func seoImageQuality(l *checkList, s *PageSignals) {
	images := s.Tags["img"]
	if len(images) == 0 {
		for _, id := range []string{"img_alt_quality", "img_alt_length", "img_lazy_loading", "img_modern_format"} {
			l.skip(id, "No images to check")
		}
		return
	}
	placeholder, long, modern, recognised, eager := 0, 0, 0, 0, 0
	for _, img := range images {
		alt := strings.TrimSpace(img.Attrs["alt"])
		if alt != "" && altFilenameRegex.MatchString(alt) {
			placeholder++
		}
		if len(alt) > 125 {
			long++
		}
		src := img.Attrs["src"]
		if cut := strings.IndexAny(src, "?#"); cut >= 0 {
			src = src[:cut]
		}
		if imageExtRegex.MatchString(src) {
			recognised++
			if modernImageExtRegex.MatchString(src) {
				modern++
			}
		}
		if !strings.EqualFold(strings.TrimSpace(img.Attrs["loading"]), "lazy") && !strings.EqualFold(strings.TrimSpace(img.Attrs["fetchpriority"]), "high") {
			eager++
		}
	}
	l.add("img_alt_quality", fmt.Sprintf("%d image alt text(s) look like filenames or placeholders", placeholder), placeholder == 0)
	l.add("img_alt_length", fmt.Sprintf("%d image alt text(s) longer than 125 characters", long), long == 0)
	l.add("img_lazy_loading", fmt.Sprintf("%d image(s) load eagerly (limit %d)", eager, maxEagerImages), eager <= maxEagerImages)
	l.add("img_modern_format", fmt.Sprintf("%d of %d images use WebP, AVIF or SVG", modern, recognised), recognised == 0 || modern*2 >= recognised)
}

func resolveNavigable(s *PageSignals, href string) *url.URL {
	href = strings.TrimSpace(href)
	if href == "" {
		return nil
	}
	for _, prefix := range nonNavigableHrefPrefixes {
		if strings.HasPrefix(href, prefix) {
			return nil
		}
	}
	parsed, err := url.Parse(href)
	if err != nil || s.URL == nil {
		return nil
	}
	return s.URL.ResolveReference(parsed)
}

func seoLinkQuality(l *checkList, s *PageSignals) {
	anchors := s.Tags["a"]
	nofollowInternal, generic, empty, withHref, javascript := 0, 0, 0, 0, 0
	for _, a := range anchors {
		href := strings.TrimSpace(a.Attrs["href"])
		if href == "" {
			continue
		}
		withHref++
		if strings.HasPrefix(strings.ToLower(href), "javascript:") {
			javascript++
		}
		if target := resolveNavigable(s, href); target != nil && target.Host == s.URL.Host && lowerContainsAny(a.Attrs["rel"], "nofollow") {
			nofollowInternal++
		}
		text := strings.ToLower(strings.Join(strings.Fields(a.Text), " "))
		if genericAnchors[text] {
			generic++
		}
		if href != "#" && text == "" && !a.HasMedia && strings.TrimSpace(a.Attrs["aria-label"]) == "" && strings.TrimSpace(a.Attrs["title"]) == "" {
			empty++
		}
	}
	hasLinks, noLinks := withHref > 0, "No links on the page"
	skipUnless(l, hasLinks, "nofollow_internal", fmt.Sprintf("%d internal link(s) marked rel=nofollow", nofollowInternal), noLinks, nofollowInternal == 0)
	skipUnless(l, hasLinks, "generic_anchor_text", fmt.Sprintf("%d link(s) use generic anchor text like \"click here\"", generic), noLinks, generic == 0)
	skipUnless(l, hasLinks, "empty_anchor_text", fmt.Sprintf("%d link(s) have no text, label or image", empty), noLinks, empty == 0)
	skipUnless(l, hasLinks, "javascript_hrefs", fmt.Sprintf("%d link(s) use javascript: hrefs", javascript), noLinks, javascript == 0)
	skipUnless(l, hasLinks, "links_count", fmt.Sprintf("%d links on page (200 or fewer recommended)", withHref), noLinks, withHref <= 200)
	seoMixedContent(l, s)
}

func seoMixedContent(l *checkList, s *PageSignals) {
	if s.URL == nil || s.URL.Scheme != schemeHTTPS {
		l.skip("mixed_content", "Page is not served over HTTPS, no mixed content to check")
		return
	}
	insecure := 0
	for _, name := range []string{"img", tagScript, "iframe", "video", "audio", "source"} {
		for _, t := range s.Tags[name] {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(t.Attrs["src"])), "http://") {
				insecure++
			}
		}
	}
	for _, t := range linkTags(s, "stylesheet") {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(t.Attrs["href"])), "http://") {
			insecure++
		}
	}
	l.add("mixed_content", fmt.Sprintf("%d insecure http:// resource(s) on an HTTPS page", insecure), insecure == 0)
}

func seoLoadPerformance(l *checkList, s *PageSignals) {
	blocking, externalScripts := 0, 0
	for _, t := range s.Tags[tagScript] {
		if strings.TrimSpace(t.Attrs["src"]) == "" {
			continue
		}
		externalScripts++
		_, async := t.Attrs["async"]
		_, deferred := t.Attrs["defer"]
		if t.InHead && !async && !deferred && !strings.EqualFold(strings.TrimSpace(t.Attrs["type"]), "module") {
			blocking++
		}
	}
	noScripts := "No external scripts"
	skipUnless(l, externalScripts > 0, "render_blocking_scripts", fmt.Sprintf("%d render-blocking script(s) in head", blocking), noScripts, blocking == 0)
	skipUnless(l, externalScripts > 0, "script_count", fmt.Sprintf("%d external scripts (20 or fewer recommended)", externalScripts), noScripts, externalScripts <= 20)

	stylesheets := linkTags(s, "stylesheet")
	skipUnless(l, len(stylesheets) > 0, "stylesheet_count", fmt.Sprintf("%d stylesheets (6 or fewer recommended)", len(stylesheets)), "No stylesheets", len(stylesheets) <= 6)

	cssBytes := 0
	for _, t := range s.Tags["style"] {
		cssBytes += len(t.Text)
	}
	l.add("inline_css_size", fmt.Sprintf("%d bytes of inline CSS (20000 or fewer recommended)", cssBytes), cssBytes <= 20000)

	l.add("resource_hints", "Preconnect or preload hints cover third-party resources", hasResourceHints(s, stylesheets))
	l.add("html_weight", fmt.Sprintf("HTML is %d bytes (500000 or fewer recommended)", len(s.RawHTML)), len(s.RawHTML) <= 500000)
}

func hasResourceHints(s *PageSignals, stylesheets []TagInfo) bool {
	for _, rel := range []string{"preconnect", "preload", "dns-prefetch", "prefetch", "modulepreload"} {
		if len(linkTags(s, rel)) > 0 {
			return true
		}
	}
	if s.URL == nil {
		return true
	}
	var resources []string
	for _, t := range s.Tags[tagScript] {
		resources = append(resources, t.Attrs["src"])
	}
	for _, t := range stylesheets {
		resources = append(resources, t.Attrs["href"])
	}
	for _, ref := range resources {
		if target := resolveNavigable(s, ref); target != nil && target.Host != s.URL.Host {
			return false
		}
	}
	return true
}

func seoAccessibility(l *checkList, s *PageSignals) {
	l.add("skip_link", "Skip-to-content link present", s.HasSkipLink)

	unnamed := 0
	for _, b := range s.Tags["button"] {
		if strings.TrimSpace(b.Text) == "" && !b.HasMedia && strings.TrimSpace(b.Attrs["aria-label"]) == "" && strings.TrimSpace(b.Attrs["title"]) == "" {
			unnamed++
		}
	}
	skipUnless(l, len(s.Tags["button"]) > 0, "button_names", fmt.Sprintf("%d button(s) have no accessible name", unnamed), "No buttons", unnamed == 0)

	headers := len(s.Tags["th"])
	skipUnless(l, len(s.Tags["table"]) > 0, "table_headers", fmt.Sprintf("%d header cell(s) for %d table(s)", headers, len(s.Tags["table"])), "No tables", headers >= len(s.Tags["table"]))

	untitled := 0
	for _, f := range s.Tags["iframe"] {
		if strings.TrimSpace(f.Attrs["title"]) == "" {
			untitled++
		}
	}
	skipUnless(l, len(s.Tags["iframe"]) > 0, "iframe_title", fmt.Sprintf("%d iframe(s) missing a title attribute", untitled), "No iframes", untitled == 0)

	seoColorContrast(l, s)
}
