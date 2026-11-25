package audit

import (
	"fmt"
	"strings"
)

const schemaIDKey = "@id"

// entityRefKeys are schema properties that point at another entity. A value
// that is only an @id, or a bare URI string, must resolve inside the graph.
var entityRefKeys = map[string]bool{
	"about": true, "author": true, "brand": true, "copyrightHolder": true,
	"creator": true, "funder": true, "isPartOf": true, "mainEntity": true,
	"member": true, "parentOrganization": true, "provider": true,
	"publisher": true, "worksFor": true,
}

type idGraph struct {
	defined map[string]bool
	refs    []string
}

func geoSchemaIDGraph(l *checkList, s *PageSignals) {
	graph := collectIDGraph(s)
	if len(graph.refs) == 0 {
		l.skip("geo_schema_id_graph_coherence", "No @id entity references to resolve")
		return
	}
	missing := missingIDRefs(graph)
	label := fmt.Sprintf("%d of %d @id reference(s) resolve to a declared node", len(graph.refs)-len(missing), len(graph.refs))
	var ev []Evidence
	for _, id := range missing {
		ev = append(ev, evidence(schemaIDKey, id))
	}
	l.addEv("geo_schema_id_graph_coherence", label, len(missing) == 0, ev...)
}

func collectIDGraph(s *PageSignals) idGraph {
	graph := idGraph{defined: map[string]bool{}}
	doc := parseSchemaDoc(s)
	for _, root := range doc.roots {
		walkIDGraph(root, &graph, false)
	}
	return graph
}

func walkIDGraph(v any, graph *idGraph, refContext bool) {
	switch n := v.(type) {
	case []any:
		for _, item := range n {
			walkIDGraph(item, graph, refContext)
		}
	case map[string]any:
		recordIDNode(n, graph)
		for key, child := range n {
			if key == schemaIDKey || key == "@context" || key == "@type" {
				continue
			}
			walkIDGraph(child, graph, entityRefKeys[key])
		}
	case string:
		if refContext {
			recordURIRef(n, graph)
		}
	}
}

func recordIDNode(node map[string]any, graph *idGraph) {
	id, _ := node[schemaIDKey].(string)
	id = strings.TrimSpace(id)
	if id == "" {
		return
	}
	if idOnlyNode(node) {
		graph.refs = append(graph.refs, id)
		return
	}
	graph.defined[id] = true
}

func idOnlyNode(node map[string]any) bool {
	for key := range node {
		if key != schemaIDKey {
			return false
		}
	}
	return true
}

func recordURIRef(raw string, graph *idGraph) {
	id := strings.TrimSpace(raw)
	if id == "" {
		return
	}
	if strings.HasPrefix(id, "#") || strings.Contains(id, "://") {
		graph.refs = append(graph.refs, id)
	}
}

func missingIDRefs(graph idGraph) []string {
	seen := map[string]bool{}
	var missing []string
	for _, id := range graph.refs {
		if graph.defined[id] || seen[id] {
			continue
		}
		seen[id] = true
		missing = append(missing, id)
	}
	return missing
}

type brandSignal struct {
	source string
	name   string
}

func geoBrandEntityConsistency(l *checkList, s *PageSignals) {
	signals := identityBrandSignals(s)
	parts := titleBrandParts(s.Title)
	if len(signals) == 0 || (len(signals) < 2 && len(parts) == 0) {
		l.skip("geo_brand_entity_consistency", "Need at least two brand signals to compare")
		return
	}
	ev := disagreeingBrands(signals)
	if len(parts) > 0 && !titlePartAgrees(parts, signals) {
		ev = append(ev, evidence("title", s.Title))
	}
	l.addEv("geo_brand_entity_consistency", fmt.Sprintf("%d brand signal(s) use a consistent name", len(signals)), len(ev) == 0, ev...)
}

func identityBrandSignals(s *PageSignals) []brandSignal {
	var out []brandSignal
	if name := strings.TrimSpace(s.OGTags["og:site_name"]); name != "" {
		out = append(out, brandSignal{"og:site_name", name})
	}
	doc := parseSchemaDoc(s)
	out = append(out, schemaBrandSignals(doc.ofType("website"), "WebSite.name")...)
	out = append(out, schemaBrandSignals(doc.ofType(append(append([]string{}, orgTypes...), localTypes...)...), "Organization.name")...)
	return out
}

func disagreeingBrands(signals []brandSignal) []Evidence {
	var ev []Evidence
	for i := range signals {
		for j := i + 1; j < len(signals); j++ {
			if brandNamesDisagree(signals[i].name, signals[j].name) {
				ev = append(ev, evidence(signals[i].source+" vs "+signals[j].source, signals[i].name+" / "+signals[j].name))
			}
		}
	}
	return ev
}

func titlePartAgrees(parts []string, signals []brandSignal) bool {
	for _, part := range parts {
		if partAgreesWithAll(part, signals) {
			return true
		}
	}
	return false
}

func partAgreesWithAll(part string, signals []brandSignal) bool {
	for _, signal := range signals {
		if brandNamesDisagree(part, signal.name) {
			return false
		}
	}
	return true
}

func schemaBrandSignals(nodes []map[string]any, source string) []brandSignal {
	seen := map[string]bool{}
	var out []brandSignal
	for _, node := range nodes {
		name, _ := node["name"].(string)
		name = strings.TrimSpace(name)
		if name == "" || seen[strings.ToLower(name)] {
			continue
		}
		seen[strings.ToLower(name)] = true
		out = append(out, brandSignal{source, name})
	}
	return out
}

// titleBrandParts splits a title on the first site-name separator. Headline
// titles without a separator are page topics, not brand names.
func titleBrandParts(title string) []string {
	for _, sep := range []string{" | ", " — ", " – ", " - "} {
		left, right, ok := strings.Cut(title, sep)
		if !ok {
			continue
		}
		left, right = strings.TrimSpace(left), strings.TrimSpace(right)
		if left == "" || right == "" {
			return nil
		}
		return []string{left, right}
	}
	return nil
}

func brandNamesDisagree(a, b string) bool {
	a, b = normBrand(a), normBrand(b)
	if a == "" || b == "" || a == b {
		return false
	}
	return !strings.Contains(a, b) && !strings.Contains(b, a)
}

func normBrand(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}
