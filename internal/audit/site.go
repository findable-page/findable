package audit

import (
	"regexp"
	"strings"
)

func crawlerCheck(robotsText, pagePath, bot string) Check {
	return Check{
		ID:    "robots-" + strings.ToLower(bot),
		Label: crawlerLabel(bot),
		Pass:  RobotsAllows(robotsText, bot, pagePath),
	}
}

func crawlerLabel(bot string) string {
	if bot == "Google-Extended" {
		return "robots.txt lets Google-Extended fetch this page (Gemini training only; it does not control AI Overviews)"
	}
	b, ok := BotByName(bot)
	if !ok {
		return "robots.txt lets " + bot + " fetch this page"
	}
	var verb string
	switch b.Purpose {
	case PurposeSearch:
		verb = "fetch this page (AI search retrieval)"
	case PurposeUserFetch:
		verb = "fetch pages users paste on demand"
	default:
		verb = "fetch this page (model training; blocking it does not affect AI search citations)"
	}
	return "robots.txt lets " + bot + " " + verb
}

const (
	pageCrawler       = "*"
	maxRobotsWildcard = 10
)

type robotsRule struct {
	allow   bool
	pattern string
}

type robotsGroup struct {
	agents []string
	rules  []robotsRule
}

func parseRobots(robots string) []robotsGroup {
	var groups []robotsGroup
	inAgents := false
	for _, line := range strings.Split(robots, "\n") {
		line, _, _ = strings.Cut(line, "#")
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		switch key {
		case "user-agent":
			if !inAgents {
				groups = append(groups, robotsGroup{})
			}
			g := &groups[len(groups)-1]
			token, _, _ := strings.Cut(value, "/")
			g.agents = append(g.agents, strings.ToLower(strings.TrimSpace(token)))
			inAgents = true
		case "allow", "disallow":
			inAgents = false
			if len(groups) > 0 && value != "" {
				g := &groups[len(groups)-1]
				g.rules = append(g.rules, robotsRule{allow: key == "allow", pattern: value})
			}
		}
	}
	return groups
}

func rulesFor(groups []robotsGroup, agent string) []robotsRule {
	agent = strings.ToLower(agent)
	var named, wildcard []robotsRule
	matchedNamed := false
	for _, g := range groups {
		for _, a := range g.agents {
			switch a {
			case agent:
				named = append(named, g.rules...)
				matchedNamed = true
			case pageCrawler:
				wildcard = append(wildcard, g.rules...)
			}
		}
	}
	if matchedNamed {
		return named
	}
	return wildcard
}

func robotsPatternMatches(pattern, path string) bool {
	if strings.Count(pattern, "*") > maxRobotsWildcard {
		return false
	}
	anchored := strings.HasSuffix(pattern, "$")
	expr := "^" + strings.ReplaceAll(regexp.QuoteMeta(strings.TrimSuffix(pattern, "$")), `\*`, ".*")
	if anchored {
		expr += "$"
	}
	re, err := regexp.Compile(expr)
	return err == nil && re.MatchString(path)
}

func RobotsAllows(robots, agent, path string) bool {
	best := robotsRule{allow: true}
	bestLen := -1
	for _, r := range rulesFor(parseRobots(robots), agent) {
		if !robotsPatternMatches(r.pattern, path) {
			continue
		}
		if len(r.pattern) > bestLen || (len(r.pattern) == bestLen && r.allow) {
			best, bestLen = r, len(r.pattern)
		}
	}
	return best.allow
}

func SiteCategory(pagePath string, robots, llms *string) Category {
	robotsText := ""
	if robots != nil {
		robotsText = *robots
	}
	checks := make([]Check, 0, len(aiSearchCrawlers)+len(trainingCrawlers)+3)
	checks = append(checks,
		Check{ID: "robots-txt", Label: "robots.txt is present", Pass: robots != nil},
		Check{ID: "robots-page", Label: "robots.txt lets crawlers fetch this page", Pass: RobotsAllows(robotsText, pageCrawler, pagePath)},
	)
	for _, bot := range aiSearchCrawlers {
		checks = append(checks, crawlerCheck(robotsText, pagePath, bot))
	}
	for _, bot := range userFetchCrawlers {
		checks = append(checks, crawlerCheck(robotsText, pagePath, bot))
	}
	for _, bot := range trainingCrawlers {
		checks = append(checks, crawlerCheck(robotsText, pagePath, bot))
	}
	checks = append(checks, Check{ID: "llms-txt", Label: "/llms.txt is present", Pass: llms != nil})
	attachMeta(checks)
	passed := 0
	for _, c := range checks {
		if c.Pass {
			passed++
		}
	}
	return Category{Passed: passed, Total: len(checks), Checks: checks}
}
