package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"

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

var statRegex = regexp.MustCompile(`(?:\$\s*\d+(?:\.\d+)?|\b\d+(?:\.\d+)?\s*(?:%|ms|s|MB|GB|KB|px|M|B|k|users|queries|hours|pages)\b)`)
var yearRegex = regexp.MustCompile(`\b202[4-9]\b`)

func extractSignals(targetURL string, rawHTML string, headers http.Header, statusCode int) PageSignals {
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

	trimmed := strings.TrimSpace(rawHTML)
	if strings.HasPrefix(strings.ToLower(trimmed), "<!doctype html") {
		sig.HTMLDoctype = true
	}

	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return sig
	}

	seenIDs := make(map[string]int)

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		sig.DOMSize++

		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)

			for _, attr := range n.Attr {
				if strings.ToLower(attr.Key) == "id" {
					val := strings.TrimSpace(attr.Val)
					if val != "" {
						seenIDs[val]++
						if seenIDs[val] == 2 {
							sig.DuplicateIDs++
						}
					}
				}
			}

			switch tag {
			case "html":
				for _, attr := range n.Attr {
					if strings.ToLower(attr.Key) == "lang" {
						sig.Lang = strings.TrimSpace(attr.Val)
					}
				}
			case "title":
				rawTitle := strings.TrimSpace(collectText(n))
				if idx := strings.Index(rawTitle, "</"); idx != -1 {
					rawTitle = strings.TrimSpace(rawTitle[:idx])
				}
				sig.Title = rawTitle
			case "meta":
				var name, property, content, httpEquiv, charset string
				for _, attr := range n.Attr {
					k := strings.ToLower(attr.Key)
					v := strings.TrimSpace(attr.Val)
					switch k {
					case "name":
						name = strings.ToLower(v)
					case "property":
						property = strings.ToLower(v)
					case "content":
						content = v
					case "http-equiv":
						httpEquiv = strings.ToLower(v)
					case "charset":
						charset = strings.ToLower(v)
					}
				}
				if charset != "" {
					sig.Charset = charset
				}
				if httpEquiv == "refresh" {
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
			case "link":
				var rel, href, sizes, linkType string
				for _, attr := range n.Attr {
					k := strings.ToLower(attr.Key)
					v := strings.TrimSpace(attr.Val)
					switch k {
					case "rel":
						rel = strings.ToLower(v)
					case "href":
						href = v
					case "sizes":
						sizes = v
					case "type":
						linkType = strings.ToLower(v)
					}
				}
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
				_ = sizes
			case "h1", "h2", "h3", "h4", "h5", "h6":
				lvl := int(tag[1] - '0')
				txt := strings.TrimSpace(collectText(n))
				sig.Headings = append(sig.Headings, Heading{Level: lvl, Tag: tag, Text: txt})
				if lvl == 1 {
					sig.H1Count++
				} else if lvl == 2 {
					sig.H2Count++
				}
			case "p":
				txt := strings.TrimSpace(collectText(n))
				if txt != "" {
					sig.Paragraphs = append(sig.Paragraphs, txt)
				}
			case "img":
				var src, alt, width, height string
				for _, attr := range n.Attr {
					k := strings.ToLower(attr.Key)
					switch k {
					case "src":
						src = attr.Val
					case "alt":
						alt = attr.Val
					case "width":
						width = attr.Val
					case "height":
						height = attr.Val
					}
				}
				sig.Images = append(sig.Images, ImageSignal{Src: src, Alt: alt, Width: width, Height: height})
			case "a":
				var href, target, rel string
				for _, attr := range n.Attr {
					k := strings.ToLower(attr.Key)
					switch k {
					case "href":
						href = strings.TrimSpace(attr.Val)
					case "target":
						target = strings.ToLower(strings.TrimSpace(attr.Val))
					case "rel":
						rel = strings.ToLower(strings.TrimSpace(attr.Val))
					}
				}
				if href == "" || href == "#" {
					sig.EmptyHrefs++
				}
				if target == "_blank" {
					sig.TargetBlankCount++
					if !strings.Contains(rel, "noopener") && !strings.Contains(rel, "noreferrer") {
						sig.MissingOpenerCount++
					}
				}
				if href != "" && !strings.HasPrefix(href, "#") && !strings.HasPrefix(href, "javascript:") && !strings.HasPrefix(href, "mailto:") && !strings.HasPrefix(href, "tel:") {
					linkURL, err := url.Parse(href)
					if err == nil {
						resolved := linkURL
						if u != nil {
							resolved = u.ResolveReference(linkURL)
						}
						resStr := resolved.String()
						if u != nil && (resolved.Host == u.Host || resolved.Host == "") {
							sig.InternalLinks = append(sig.InternalLinks, resStr)
						} else {
							sig.ExternalLinks = append(sig.ExternalLinks, resStr)
						}
					}
				}
				if strings.Contains(strings.ToLower(collectText(n)), "skip to") {
					sig.HasSkipLink = true
				}
			case "article", "section", "main", "nav", "aside", "header", "footer", "time", "figure", "details":
				sig.SemanticCounts[tag]++
				if tag == "details" {
					sig.DetailsCount++
				}
			case "dl":
				sig.DLCount++
			case "ol":
				sig.OLCount++
			case "ul":
				sig.ULCount++
			case "table":
				sig.TableCount++
			case "code", "pre":
				sig.CodeCount++
			case "blockquote":
				sig.BlockquoteCount++
				quoteText := strings.TrimSpace(collectText(n))
				if quoteText != "" {
					sig.ExpertQuotes = append(sig.ExpertQuotes, quoteText)
				}
			case "input", "textarea", "select":
				sig.FormInputCount++
			case "script":
				isJSONLD := false
				for _, attr := range n.Attr {
					if strings.ToLower(attr.Key) == "type" && strings.ToLower(strings.TrimSpace(attr.Val)) == "application/ld+json" {
						isJSONLD = true
						break
					}
				}
				if isJSONLD {
					parseJSONLD(collectRawText(n), &sig)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	// Collect full visible text content
	sig.TextContent = extractVisibleText(doc)
	fields := strings.Fields(sig.TextContent)
	sig.Words = fields

	matches := statRegex.FindAllString(sig.TextContent, -1)
	sig.NumbersWithUnits = matches

	return sig
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

	var inspect func(any)
	inspect = func(item any) {
		switch v := item.(type) {
		case map[string]any:
			sig.JSONLDs = append(sig.JSONLDs, v)
			if t, ok := v["@type"].(string); ok {
				sig.SchemaTypes = append(sig.SchemaTypes, t)
			}
			if graph, ok := v["@graph"].([]any); ok {
				for _, g := range graph {
					inspect(g)
				}
			}
			if sameAs, ok := v["sameAs"]; ok {
				switch sa := sameAs.(type) {
				case string:
					sig.SameAsLinks = append(sig.SameAsLinks, sa)
				case []any:
					for _, s := range sa {
						if str, ok := s.(string); ok {
							sig.SameAsLinks = append(sig.SameAsLinks, str)
						}
					}
				}
			}
			if auth, ok := v["author"].(map[string]any); ok {
				if name, ok := auth["name"].(string); ok {
					sig.AuthorInSchema = name
				}
			}
			if dp, ok := v["datePublished"].(string); ok {
				sig.DatePublished = dp
			}
			if dm, ok := v["dateModified"].(string); ok {
				sig.DateModified = dm
			}
		case []any:
			for _, sub := range v {
				inspect(sub)
			}
		}
	}
	inspect(root)
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

func containsWord(text, word string) bool {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	w := strings.ToLower(word)
	for _, f := range fields {
		if f == w {
			return true
		}
	}
	return false
}
