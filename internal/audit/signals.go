package audit

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type Heading struct {
	Level int
	Tag   string
	Text  string
}

type ImageSignal struct {
	Src    string
	Alt    string
	Width  string
	Height string
}

type PageSignals struct {
	URL                 *url.URL
	TargetURL           string
	RawHTML             string
	StatusCode          int
	Headers             http.Header
	HTMLDoctype         bool
	Title               string
	MetaDesc            string
	Canonical           string
	RobotsMeta          string
	Viewport            string
	Charset             string
	Lang                string
	ThemeColor          string
	MetaRefresh         string
	Headings            []Heading
	H1Count             int
	H2Count             int
	Paragraphs          []string
	Words               []string
	TextContent         string
	Images              []ImageSignal
	InternalLinks       []string
	ExternalLinks       []string
	EmptyHrefs          int
	TargetBlankCount    int
	MissingOpenerCount  int
	SemanticCounts      map[string]int
	DLCount             int
	OLCount             int
	ULCount             int
	TableCount          int
	CodeCount           int
	BlockquoteCount     int
	DetailsCount        int
	JSONLDs             []map[string]any
	SchemaTypes         []string
	SameAsLinks         []string
	AuthorInSchema      string
	DatePublished       string
	DateModified        string
	OGTags              map[string]string
	TwitterTags         map[string]string
	NumbersWithUnits    []string
	ExpertQuotes        []string
	DOMSize             int
	HasSkipLink         bool
	FormInputCount      int
	UnlabeledInputCount int
	DuplicateIDs        int
	Favicon             string
	WebManifest         string
	RSSFeed             string
}

var (
	statRegex = regexp.MustCompile(`(?:\$\s*\d+(?:\.\d+)?|\b\d+(?:\.\d+)?\s*(?:%|ms|s|MB|GB|KB|px|M|B|k|users|queries|hours|pages)\b)`)
	yearRegex = regexp.MustCompile(`\b202[4-9]\b`)
)

func extractSignals(targetURL, rawHTML string, headers http.Header, statusCode int) PageSignals {
	u, _ := url.Parse(targetURL)
	sig := PageSignals{
		URL:            u,
		TargetURL:      targetURL,
		RawHTML:        rawHTML,
		StatusCode:     statusCode,
		Headers:        headers,
		SemanticCounts: make(map[string]int),
		OGTags:         make(map[string]string),
		TwitterTags:    make(map[string]string),
	}
	if headers == nil {
		sig.Headers = make(http.Header)
	}
	sig.HTMLDoctype = strings.HasPrefix(strings.ToLower(strings.TrimSpace(rawHTML)), "<!doctype html")

	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return sig
	}

	seenIDs := make(map[string]int)
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		sig.DOMSize++
		if n.Type == html.ElementNode {
			countDuplicateIDs(&sig, n, seenIDs)
			if collect, ok := elementCollectors[strings.ToLower(n.Data)]; ok {
				collect(&sig, n)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	sig.TextContent = extractVisibleText(doc)
	sig.Words = strings.Fields(sig.TextContent)
	sig.NumbersWithUnits = statRegex.FindAllString(sig.TextContent, -1)
	return sig
}

var elementCollectors = map[string]func(*PageSignals, *html.Node){
	"html":       collectLang,
	"title":      collectTitle,
	"meta":       collectMeta,
	"link":       collectLinkTag,
	"h1":         collectHeading,
	"h2":         collectHeading,
	"h3":         collectHeading,
	"h4":         collectHeading,
	"h5":         collectHeading,
	"h6":         collectHeading,
	"p":          collectParagraph,
	"img":        collectImage,
	"a":          collectAnchor,
	"article":    countSemantic,
	"section":    countSemantic,
	"main":       countSemantic,
	"nav":        countSemantic,
	"aside":      countSemantic,
	"header":     countSemantic,
	"footer":     countSemantic,
	"time":       countSemantic,
	"figure":     countSemantic,
	"details":    countSemantic,
	"dl":         func(s *PageSignals, _ *html.Node) { s.DLCount++ },
	"ol":         func(s *PageSignals, _ *html.Node) { s.OLCount++ },
	"ul":         func(s *PageSignals, _ *html.Node) { s.ULCount++ },
	"table":      func(s *PageSignals, _ *html.Node) { s.TableCount++ },
	"code":       func(s *PageSignals, _ *html.Node) { s.CodeCount++ },
	"pre":        func(s *PageSignals, _ *html.Node) { s.CodeCount++ },
	"input":      func(s *PageSignals, _ *html.Node) { s.FormInputCount++ },
	"textarea":   func(s *PageSignals, _ *html.Node) { s.FormInputCount++ },
	"select":     func(s *PageSignals, _ *html.Node) { s.FormInputCount++ },
	"blockquote": collectBlockquote,
	"script":     collectScript,
}

func attrs(n *html.Node) map[string]string {
	out := make(map[string]string, len(n.Attr))
	for _, a := range n.Attr {
		out[strings.ToLower(a.Key)] = a.Val
	}
	return out
}

func trimmedAttr(a map[string]string, key string) string {
	return strings.TrimSpace(a[key])
}

func countDuplicateIDs(sig *PageSignals, n *html.Node, seen map[string]int) {
	for _, a := range n.Attr {
		if !strings.EqualFold(a.Key, "id") {
			continue
		}
		id := strings.TrimSpace(a.Val)
		if id == "" {
			continue
		}
		seen[id]++
		if seen[id] == 2 {
			sig.DuplicateIDs++
		}
	}
}

func collectLang(sig *PageSignals, n *html.Node) {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, "lang") {
			sig.Lang = strings.TrimSpace(a.Val)
		}
	}
}

func collectTitle(sig *PageSignals, n *html.Node) {
	title := strings.TrimSpace(collectText(n))
	if idx := strings.Index(title, "</"); idx != -1 {
		title = strings.TrimSpace(title[:idx])
	}
	sig.Title = title
}

func collectMeta(sig *PageSignals, n *html.Node) {
	a := attrs(n)
	name := strings.ToLower(trimmedAttr(a, "name"))
	property := strings.ToLower(trimmedAttr(a, "property"))
	content := trimmedAttr(a, "content")
	if charset := strings.ToLower(trimmedAttr(a, "charset")); charset != "" {
		sig.Charset = charset
	}
	if strings.ToLower(trimmedAttr(a, "http-equiv")) == "refresh" {
		sig.MetaRefresh = content
	}
	switch name {
	case "description":
		sig.MetaDesc = content
	case "viewport":
		sig.Viewport = content
	case "robots":
		sig.RobotsMeta = content
	case "theme-color":
		sig.ThemeColor = content
	}
	if strings.HasPrefix(name, "twitter:") {
		sig.TwitterTags[name] = content
	}
	if strings.HasPrefix(property, "og:") {
		sig.OGTags[property] = content
	}
}

func collectLinkTag(sig *PageSignals, n *html.Node) {
	a := attrs(n)
	rel := strings.ToLower(trimmedAttr(a, "rel"))
	href := trimmedAttr(a, "href")
	linkType := strings.ToLower(trimmedAttr(a, "type"))
	if rel == "canonical" {
		sig.Canonical = href
	}
	if strings.Contains(rel, "icon") {
		sig.Favicon = href
	}
	if rel == "manifest" {
		sig.WebManifest = href
	}
	if rel == "alternate" && (strings.Contains(linkType, "rss") || strings.Contains(linkType, "atom")) {
		sig.RSSFeed = href
	}
}

func collectHeading(sig *PageSignals, n *html.Node) {
	tag := strings.ToLower(n.Data)
	level := int(tag[1] - '0')
	sig.Headings = append(sig.Headings, Heading{Level: level, Tag: tag, Text: strings.TrimSpace(collectText(n))})
	switch level {
	case 1:
		sig.H1Count++
	case 2:
		sig.H2Count++
	}
}

func collectParagraph(sig *PageSignals, n *html.Node) {
	if text := strings.TrimSpace(collectText(n)); text != "" {
		sig.Paragraphs = append(sig.Paragraphs, text)
	}
}

func collectImage(sig *PageSignals, n *html.Node) {
	a := attrs(n)
	sig.Images = append(sig.Images, ImageSignal{Src: a["src"], Alt: a["alt"], Width: a["width"], Height: a["height"]})
}

func collectAnchor(sig *PageSignals, n *html.Node) {
	a := attrs(n)
	href := trimmedAttr(a, "href")
	rel := strings.ToLower(trimmedAttr(a, "rel"))
	if href == "" || href == "#" {
		sig.EmptyHrefs++
	}
	if strings.ToLower(trimmedAttr(a, "target")) == "_blank" {
		sig.TargetBlankCount++
		if !strings.Contains(rel, "noopener") && !strings.Contains(rel, "noreferrer") {
			sig.MissingOpenerCount++
		}
	}
	classifyLink(sig, href)
	if strings.Contains(strings.ToLower(collectText(n)), "skip to") {
		sig.HasSkipLink = true
	}
}

var nonNavigableHrefPrefixes = []string{"#", "javascript:", "mailto:", "tel:"}

func classifyLink(sig *PageSignals, href string) {
	if href == "" {
		return
	}
	for _, prefix := range nonNavigableHrefPrefixes {
		if strings.HasPrefix(href, prefix) {
			return
		}
	}
	linkURL, err := url.Parse(href)
	if err != nil {
		return
	}
	base := sig.URL
	resolved := linkURL
	if base != nil {
		resolved = base.ResolveReference(linkURL)
	}
	if base != nil && (resolved.Host == base.Host || resolved.Host == "") {
		sig.InternalLinks = append(sig.InternalLinks, resolved.String())
	} else {
		sig.ExternalLinks = append(sig.ExternalLinks, resolved.String())
	}
}

func countSemantic(sig *PageSignals, n *html.Node) {
	tag := strings.ToLower(n.Data)
	sig.SemanticCounts[tag]++
	if tag == "details" {
		sig.DetailsCount++
	}
}

func collectBlockquote(sig *PageSignals, n *html.Node) {
	sig.BlockquoteCount++
	if quote := strings.TrimSpace(collectText(n)); quote != "" {
		sig.ExpertQuotes = append(sig.ExpertQuotes, quote)
	}
}

func collectScript(sig *PageSignals, n *html.Node) {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, "type") && strings.EqualFold(strings.TrimSpace(a.Val), "application/ld+json") {
			parseJSONLD(collectRawText(n), sig)
			return
		}
	}
}

func parseJSONLD(raw string, sig *PageSignals) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}
	var root any
	if err := json.Unmarshal([]byte(raw), &root); err != nil {
		return
	}
	inspectJSONLD(root, sig)
}

func inspectJSONLD(item any, sig *PageSignals) {
	switch v := item.(type) {
	case map[string]any:
		collectSchemaNode(v, sig)
	case []any:
		for _, sub := range v {
			inspectJSONLD(sub, sig)
		}
	}
}

func collectSchemaNode(node map[string]any, sig *PageSignals) {
	sig.JSONLDs = append(sig.JSONLDs, node)
	if t, ok := node["@type"].(string); ok {
		sig.SchemaTypes = append(sig.SchemaTypes, t)
	}
	if graph, ok := node["@graph"].([]any); ok {
		for _, g := range graph {
			inspectJSONLD(g, sig)
		}
	}
	sig.SameAsLinks = append(sig.SameAsLinks, sameAsLinks(node["sameAs"])...)
	if author, ok := node["author"].(map[string]any); ok {
		if name, ok := author["name"].(string); ok {
			sig.AuthorInSchema = name
		}
	}
	if published, ok := node["datePublished"].(string); ok {
		sig.DatePublished = published
	}
	if modified, ok := node["dateModified"].(string); ok {
		sig.DateModified = modified
	}
}

func sameAsLinks(sameAs any) []string {
	switch v := sameAs.(type) {
	case string:
		return []string{v}
	case []any:
		links := make([]string, 0, len(v))
		for _, item := range v {
			if link, ok := item.(string); ok {
				links = append(links, link)
			}
		}
		return links
	}
	return nil
}

func collectText(n *html.Node) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(curr *html.Node) {
		if curr.Type == html.TextNode {
			b.WriteString(curr.Data)
		}
		for c := curr.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return b.String()
}

func collectRawText(n *html.Node) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		}
	}
	return b.String()
}

func extractVisibleText(doc *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)
			if tag == "script" || tag == "style" || tag == "noscript" || tag == "svg" {
				return
			}
		}
		if n.Type == html.TextNode {
			clean := strings.TrimSpace(n.Data)
			if clean != "" {
				b.WriteString(clean)
				b.WriteByte(' ')
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return strings.TrimSpace(b.String())
}
