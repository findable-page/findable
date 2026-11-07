package audit

import (
	"strings"

	"golang.org/x/net/html"
)

const robotsMetaName = "robots"

// XRobotsRules holds parsed robots directives per agent, with "*" as the
// wildcard. Used for both X-Robots-Tag headers and meta robots tags.
type XRobotsRules map[string]map[string]bool

// ParseXRobotsTag parses X-Robots-Tag header values, including the
// per-crawler form "X-Robots-Tag: GPTBot: noindex".
func ParseXRobotsTag(values []string) XRobotsRules {
	rules := XRobotsRules{}
	add := func(agent, tokens string) {
		set := rules[agent]
		if set == nil {
			set = map[string]bool{}
			rules[agent] = set
		}
		for _, token := range strings.Split(tokens, ",") {
			if t := strings.TrimSpace(strings.ToLower(token)); t != "" {
				set[t] = true
			}
		}
	}
	for _, header := range values {
		agent := pageCrawler
		for _, part := range strings.Split(header, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			// tokens are plain words; "agent: token" starts a new agent group
			if name, tokens, ok := strings.Cut(part, ":"); ok && strings.TrimSpace(tokens) != "" && !strings.ContainsAny(name, " /=") {
				agent = strings.ToLower(strings.TrimSpace(name))
				part = tokens
			}
			add(agent, part)
		}
	}
	return rules
}

// ParseMetaRobots extracts per-agent robots meta directives from HTML:
// <meta name="robots"> maps to "*", <meta name="GPTBot"> to "gptbot".
// Only agent-looking names are recorded.
func ParseMetaRobots(rawHTML string) XRobotsRules {
	rules := XRobotsRules{}
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return rules
	}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, "meta") {
			name, content := "", ""
			for _, a := range n.Attr {
				switch strings.ToLower(a.Key) {
				case "name":
					name = strings.ToLower(strings.TrimSpace(a.Val))
				case "content":
					content = a.Val
				}
			}
			if isRobotsAgentName(name) && strings.TrimSpace(content) != "" {
				if name == robotsMetaName {
					name = pageCrawler
				}
				set := rules[name]
				if set == nil {
					set = map[string]bool{}
					rules[name] = set
				}
				for _, token := range strings.Split(content, ",") {
					if t := strings.TrimSpace(strings.ToLower(token)); t != "" {
						set[t] = true
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return rules
}

// isRobotsAgentName accepts "robots", table bot names (lowercased) and
// bot-like suffixes for crawlers not yet in the table.
func isRobotsAgentName(name string) bool {
	if name == pageCrawler || name == robotsMetaName {
		return true
	}
	for _, b := range Bots {
		if strings.ToLower(b.Name) == name {
			return true
		}
	}
	return strings.Contains(name, "bot") || strings.Contains(name, "agent") || strings.Contains(name, "user") || strings.Contains(name, "extended")
}

// NoIndexFor reports whether the rules noindex content for the agent. A
// bot-specific group overrides the wildcard group, mirroring robots.txt
// semantics.
func (r XRobotsRules) NoIndexFor(agent string) bool {
	if set, ok := r[strings.ToLower(agent)]; ok {
		return set["noindex"]
	}
	if set, ok := r[pageCrawler]; ok {
		return set["noindex"]
	}
	return false
}
