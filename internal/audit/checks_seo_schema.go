package audit

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	isoDateRegex  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}([T ]\d{2}:\d{2}(:\d{2}(\.\d+)?)?(Z|[+-]\d{2}:?\d{2})?)?$`)
	httpURLRegex  = regexp.MustCompile(`(?i)^https?://`)
	urlSchemaKeys = map[string]bool{"url": true, "logo": true, "image": true, "sameas": true, "thumbnailurl": true, "contenturl": true, "embedurl": true}
	articleTypes  = []string{"article", "newsarticle", "blogposting", "techarticle", "scholarlyarticle"}
	orgTypes      = []string{"organization", "corporation", "ngo", "educationalorganization", "onlinestore", "newsmediaorganization"}
	localTypes    = []string{"localbusiness", "restaurant", "bakery", "store", "dentist", "hotel", "cafeorcoffeeshop", "barorpub", "medicalbusiness", "realestateagent", "autodealer", "healthandbeautybusiness"}
)

type schemaDoc struct {
	blocks  int
	invalid int
	roots   []any
	nodes   []map[string]any
}

func parseSchemaDoc(s *PageSignals) schemaDoc {
	var doc schemaDoc
	for _, script := range s.Tags[tagScript] {
		if !strings.EqualFold(strings.TrimSpace(script.Attrs["type"]), "application/ld+json") {
			continue
		}
		doc.blocks++
		var root any
		if err := json.Unmarshal([]byte(strings.TrimSpace(script.Text)), &root); err != nil {
			doc.invalid++
			continue
		}
		doc.roots = append(doc.roots, root)
		for _, item := range jsonItems(root) {
			node, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if graph, has := node["@graph"]; has {
				for _, member := range jsonItems(graph) {
					if m, isMap := member.(map[string]any); isMap {
						doc.nodes = append(doc.nodes, m)
					}
				}
			}
			if _, typed := node["@type"]; typed {
				doc.nodes = append(doc.nodes, node)
			}
		}
	}
	return doc
}

func jsonItems(v any) []any {
	switch t := v.(type) {
	case nil:
		return nil
	case []any:
		return t
	}
	return []any{v}
}

func jsonPresent(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(t) != ""
	case []any:
		return len(t) > 0
	case map[string]any:
		return len(t) > 0
	}
	return true
}

func nodeTypes(node map[string]any) []string {
	var out []string
	for _, t := range jsonItems(node["@type"]) {
		if name, ok := t.(string); ok {
			out = append(out, strings.ToLower(strings.TrimSpace(name)))
		}
	}
	return out
}

func (d schemaDoc) ofType(wanted ...string) []map[string]any {
	var out []map[string]any
	for _, node := range d.nodes {
		for _, t := range nodeTypes(node) {
			if containsString(wanted, t) {
				out = append(out, node)
				break
			}
		}
	}
	return out
}

func containsString(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

func allNodes(nodes []map[string]any, ok func(map[string]any) bool) bool {
	for _, n := range nodes {
		if !ok(n) {
			return false
		}
	}
	return true
}

func countSchemaStrings(v any, visit func(key, value string) bool) int {
	count := 0
	var walk func(key string, v any)
	walk = func(key string, v any) {
		switch t := v.(type) {
		case string:
			if visit(key, t) {
				count++
			}
		case []any:
			for _, e := range t {
				walk(key, e)
			}
		case map[string]any:
			for k, e := range t {
				walk(k, e)
			}
		}
	}
	walk("", v)
	return count
}

func contextOK(v any) bool {
	switch t := v.(type) {
	case string:
		return strings.Contains(t, "schema.org")
	case []any:
		for _, e := range t {
			if contextOK(e) {
				return true
			}
		}
	case map[string]any:
		for _, e := range t {
			if contextOK(e) {
				return true
			}
		}
	}
	return false
}

func isNumeric(v any) bool {
	switch t := v.(type) {
	case float64:
		return true
	case string:
		_, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return err == nil
	}
	return false
}

func seoSchema(l *checkList, s *PageSignals) {
	doc := parseSchemaDoc(s)
	if doc.blocks == 0 {
		for _, id := range schemaCheckIDs {
			l.skip(id, "No JSON-LD structured data to validate")
		}
		return
	}
	l.add("jsonld_valid", fmt.Sprintf("%d of %d JSON-LD block(s) are not valid JSON", doc.invalid, doc.blocks), doc.invalid == 0)
	schemaContext(l, doc)
	schemaHygiene(l, doc)
	schemaOrganization(l, doc)
	schemaWebsite(l, doc)
	schemaBreadcrumbAndFAQ(l, doc)
	schemaArticle(l, doc)
	schemaProduct(l, doc)
	schemaBusinessAndVideo(l, doc)
}

var schemaCheckIDs = []string{
	"jsonld_valid", "jsonld_context", "schema_empty_values", "schema_urls_absolute",
	"org_name", "org_logo", "org_url", "website_name_url", "searchaction_valid",
	"breadcrumb_valid", "faq_valid", "article_headline", "article_author", "article_dates", "article_image",
	"product_offers", "product_price", "product_rating", "localbusiness_address", "localbusiness_phone", "video_fields",
}

func skipUnless(l *checkList, present bool, id, label, absent string, pass bool) {
	if !present {
		l.skip(id, absent)
		return
	}
	l.add(id, label, pass)
}

func schemaContext(l *checkList, doc schemaDoc) {
	missing, roots := 0, 0
	for _, root := range doc.roots {
		for _, item := range jsonItems(root) {
			if m, ok := item.(map[string]any); ok {
				roots++
				if !contextOK(m["@context"]) {
					missing++
				}
			}
		}
	}
	skipUnless(l, roots > 0, "jsonld_context", fmt.Sprintf("%d of %d JSON-LD root(s) lack a schema.org @context", missing, roots), "No valid JSON-LD to check", missing == 0)
}

func schemaHygiene(l *checkList, doc schemaDoc) {
	empty, relative := 0, 0
	for _, root := range doc.roots {
		empty += countSchemaStrings(root, func(key, value string) bool {
			return !strings.HasPrefix(key, "@") && strings.TrimSpace(value) == ""
		})
		relative += countSchemaStrings(root, func(key, value string) bool {
			return urlSchemaKeys[strings.ToLower(key)] && strings.TrimSpace(value) != "" && !httpURLRegex.MatchString(strings.TrimSpace(value))
		})
	}
	skipUnless(l, len(doc.roots) > 0, "schema_empty_values", fmt.Sprintf("%d empty string value(s) in structured data", empty), "No valid JSON-LD to check", empty == 0)
	skipUnless(l, len(doc.roots) > 0, "schema_urls_absolute", fmt.Sprintf("%d url/logo/image value(s) are not absolute http(s) URLs", relative), "No valid JSON-LD to check", relative == 0)
}

func schemaOrganization(l *checkList, doc schemaDoc) {
	orgs := doc.ofType(append(append([]string{}, orgTypes...), localTypes...)...)
	has := func(key string) func(map[string]any) bool {
		return func(n map[string]any) bool { return jsonPresent(n[key]) }
	}
	skipUnless(l, len(orgs) > 0, "org_name", "Organization schema has a name", "No Organization schema to validate", allNodes(orgs, has("name")))
	skipUnless(l, len(orgs) > 0, "org_logo", "Organization schema has a logo", "No Organization schema to validate", allNodes(orgs, has("logo")))
	skipUnless(l, len(orgs) > 0, "org_url", "Organization schema has a url", "No Organization schema to validate", allNodes(orgs, has("url")))
}

func schemaWebsite(l *checkList, doc schemaDoc) {
	sites := doc.ofType("website")
	skipUnless(l, len(sites) > 0, "website_name_url", "WebSite schema has a name and url", "No WebSite schema to validate", allNodes(sites, func(n map[string]any) bool {
		return jsonPresent(n["name"]) && jsonPresent(n["url"])
	}))
	var searches []map[string]any
	for _, site := range sites {
		for _, action := range jsonItems(site["potentialAction"]) {
			if m, ok := action.(map[string]any); ok && containsString(nodeTypes(m), "searchaction") {
				searches = append(searches, m)
			}
		}
	}
	skipUnless(l, len(searches) > 0, "searchaction_valid", "WebSite SearchAction has a target and query-input", "No SearchAction to validate", allNodes(searches, func(n map[string]any) bool {
		return jsonPresent(n["target"]) && jsonPresent(n["query-input"])
	}))
}

func schemaBreadcrumbAndFAQ(l *checkList, doc schemaDoc) {
	crumbs := doc.ofType("breadcrumblist")
	skipUnless(l, len(crumbs) > 0, "breadcrumb_valid", "BreadcrumbList has 2+ positioned, named items", "No BreadcrumbList to validate", allNodes(crumbs, func(n map[string]any) bool {
		items := jsonItems(n["itemListElement"])
		if len(items) < 2 {
			return false
		}
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok || !isNumeric(m["position"]) {
				return false
			}
			target, _ := m["item"].(map[string]any)
			if !jsonPresent(m["name"]) && !jsonPresent(target["name"]) {
				return false
			}
		}
		return true
	}))
	faqs := doc.ofType("faqpage")
	skipUnless(l, len(faqs) > 0, "faq_valid", "FAQPage questions each have a name and answer text", "No FAQPage to validate", allNodes(faqs, func(n map[string]any) bool {
		questions := jsonItems(n["mainEntity"])
		if len(questions) == 0 {
			return false
		}
		for _, q := range questions {
			m, ok := q.(map[string]any)
			if !ok || !jsonPresent(m["name"]) {
				return false
			}
			answer, _ := m["acceptedAnswer"].(map[string]any)
			if !jsonPresent(answer["text"]) {
				return false
			}
		}
		return true
	}))
}

func schemaArticle(l *checkList, doc schemaDoc) {
	articles := doc.ofType(articleTypes...)
	absent := "No Article schema to validate"
	skipUnless(l, len(articles) > 0, "article_headline", "Article headline is present and 110 characters or fewer", absent, allNodes(articles, func(n map[string]any) bool {
		headline, _ := n["headline"].(string)
		return strings.TrimSpace(headline) != "" && utf8.RuneCountInString(headline) <= 110
	}))
	skipUnless(l, len(articles) > 0, "article_author", "Article author is named", absent, allNodes(articles, func(n map[string]any) bool {
		authors := jsonItems(n["author"])
		if len(authors) == 0 {
			return false
		}
		for _, a := range authors {
			if m, ok := a.(map[string]any); ok {
				if !jsonPresent(m["name"]) {
					return false
				}
			} else if !jsonPresent(a) {
				return false
			}
		}
		return true
	}))
	skipUnless(l, len(articles) > 0, "article_dates", "Article datePublished (and dateModified) are valid ISO dates", absent, allNodes(articles, func(n map[string]any) bool {
		published, _ := n["datePublished"].(string)
		if !isoDateRegex.MatchString(strings.TrimSpace(published)) {
			return false
		}
		if _, has := n["dateModified"]; has {
			modified, _ := n["dateModified"].(string)
			return isoDateRegex.MatchString(strings.TrimSpace(modified))
		}
		return true
	}))
	skipUnless(l, len(articles) > 0, "article_image", "Article schema has an image", absent, allNodes(articles, func(n map[string]any) bool { return jsonPresent(n["image"]) }))
}

func offerValid(v any) bool {
	m, ok := v.(map[string]any)
	if !ok {
		return false
	}
	priced := jsonPresent(m["price"]) || jsonPresent(m["lowPrice"]) || jsonPresent(m["priceSpecification"])
	return priced && (jsonPresent(m["priceCurrency"]) || jsonPresent(m["priceSpecification"]))
}

func schemaProduct(l *checkList, doc schemaDoc) {
	products := doc.ofType("product")
	absent := "No Product schema to validate"
	skipUnless(l, len(products) > 0, "product_offers", "Product schema has offers", absent, allNodes(products, func(n map[string]any) bool { return jsonPresent(n["offers"]) }))
	skipUnless(l, len(products) > 0, "product_price", "Product offers have a price and currency", absent, allNodes(products, func(n map[string]any) bool {
		offers := jsonItems(n["offers"])
		for _, o := range offers {
			if !offerValid(o) {
				return false
			}
		}
		return len(offers) > 0
	}))
	var rated []map[string]any
	for _, node := range doc.nodes {
		if rating, ok := node["aggregateRating"].(map[string]any); ok {
			rated = append(rated, rating)
		}
	}
	skipUnless(l, len(rated) > 0, "product_rating", "aggregateRating has a ratingValue and a review or rating count", "No aggregateRating to validate", allNodes(rated, func(n map[string]any) bool {
		return jsonPresent(n["ratingValue"]) && (jsonPresent(n["reviewCount"]) || jsonPresent(n["ratingCount"]))
	}))
}

func schemaBusinessAndVideo(l *checkList, doc schemaDoc) {
	local := doc.ofType(localTypes...)
	absent := "No LocalBusiness schema to validate"
	skipUnless(l, len(local) > 0, "localbusiness_address", "LocalBusiness schema has an address", absent, allNodes(local, func(n map[string]any) bool { return jsonPresent(n["address"]) }))
	skipUnless(l, len(local) > 0, "localbusiness_phone", "LocalBusiness schema has a telephone", absent, allNodes(local, func(n map[string]any) bool { return jsonPresent(n["telephone"]) }))
	videos := doc.ofType("videoobject")
	skipUnless(l, len(videos) > 0, "video_fields", "VideoObject has a name, thumbnailUrl and uploadDate", "No VideoObject to validate", allNodes(videos, func(n map[string]any) bool {
		return jsonPresent(n["name"]) && jsonPresent(n["thumbnailUrl"]) && jsonPresent(n["uploadDate"])
	}))
}

func seoSocialCompleteness(l *checkList, s *PageSignals) {
	meta := func(prop string) string { return metaProperty(s, prop) }
	l.add("og_type", "Open Graph og:type is set", meta("og:type") != "")
	l.add("og_site_name", "Open Graph og:site_name is set", meta("og:site_name") != "")
	if image := meta("og:image"); image == "" {
		l.skip("og_image_absolute", "No og:image to validate")
	} else {
		l.add("og_image_absolute", "og:image is an absolute http(s) URL", httpURLRegex.MatchString(image))
	}
	ogURL, canonical := meta("og:url"), firstHref(linkTags(s, "canonical"))
	if ogURL == "" || canonical == "" || s.URL == nil {
		l.skip("og_url_matches_canonical", "Needs both og:url and a canonical")
		return
	}
	l.add("og_url_matches_canonical", "og:url matches the canonical URL", samePage(s.URL, ogURL, canonical))
}

func samePage(base *url.URL, a, b string) bool {
	left, errA := url.Parse(a)
	right, errB := url.Parse(b)
	if errA != nil || errB != nil {
		return false
	}
	return pagePathKey(base.ResolveReference(left)) == pagePathKey(base.ResolveReference(right))
}
