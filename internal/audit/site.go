package audit

import (
	"regexp"
	"strings"
)

var aiCrawlers = []string{"GPTBot", "ClaudeBot", "PerplexityBot", "Google-Extended"}

const pageCrawler = "*"

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
			g.agents = append(g.agents, strings.ToLower(value))
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
	checks := make([]Check, 0, len(aiCrawlers)+3)
	checks = append(checks,
		Check{ID: "robots-txt", Label: "robots.txt is present", Pass: robots != nil},
		Check{ID: "robots-page", Label: "robots.txt lets crawlers fetch this page", Pass: RobotsAllows(robotsText, pageCrawler, pagePath)},
	)
	for _, bot := range aiCrawlers {
		checks = append(checks, Check{
			ID:    "robots-" + strings.ToLower(bot),
			Label: "robots.txt lets " + bot + " fetch this page",
			Pass:  RobotsAllows(robotsText, bot, pagePath),
		})
	}
	checks = append(checks, Check{ID: "llms-txt", Label: "/llms.txt is present", Pass: llms != nil})
	passed := 0
	for _, c := range checks {
		if c.Pass {
			passed++
		}
	}
	return Category{Passed: passed, Total: len(checks), Checks: checks}
}
