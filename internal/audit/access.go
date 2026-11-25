package audit

import (
	"fmt"
	"strings"
)

const accessBlockedMessage = "This page cannot be cited by AI search engines because crawler access or indexing is blocked."

// AccessGate is the Gate 1 result: whether a crawler can fetch and index the
// page. Content scores stay on the scan; this says when those scores do not
// mean the page can be cited.
type AccessGate struct {
	Gate          string         `json:"gate"`
	AccessBlocked bool           `json:"accessBlocked"`
	Status        string         `json:"status"`
	Reasons       []AccessReason `json:"reasons,omitempty"`
	Message       string         `json:"message,omitempty"`
}

// AccessReason names the check or transport fact that closed Gate 1.
type AccessReason struct {
	ID     string `json:"id"`
	Detail string `json:"detail"`
}

// RefreshAccess rebuilds the Gate 1 result from the HTTP status and the
// access checks already on the scan, including site checks when present.
func (s *Scan) RefreshAccess() {
	reasons := httpStatusReason(s.StatusCode)
	seen := map[string]bool{}
	for _, named := range s.Reported() {
		for _, check := range named.Cat.Checks {
			if gate1Failure(check) && !seen[check.ID] {
				seen[check.ID] = true
				reasons = append(reasons, AccessReason{ID: check.ID, Detail: check.Label})
			}
		}
	}
	for _, extra := range s.directiveReasons {
		if extra.ID == "" || seen[extra.ID] {
			continue
		}
		seen[extra.ID] = true
		reasons = append(reasons, extra)
	}
	gate := AccessGate{Gate: "open", Status: "OPEN"}
	if len(reasons) > 0 {
		gate.Gate = "blocked"
		gate.AccessBlocked = true
		gate.Status = "BLOCKED_FROM_INDEXING"
		gate.Reasons = reasons
		gate.Message = accessBlockedMessage
	}
	s.Access = &gate
}

func httpStatusReason(status int) []AccessReason {
	if status > 0 && (status < 200 || status >= 300) {
		return []AccessReason{{ID: "http_status", Detail: fmt.Sprintf("HTTP %d", status)}}
	}
	return nil
}

func gate1Failure(check Check) bool {
	if Status(check) != StatusFail {
		return false
	}
	switch check.ID {
	case "robots_meta", "x_robots_tag_noindex", "geo_snippet_control", checkRobotsPage:
		return true
	}
	bot, ok := strings.CutPrefix(check.ID, "robots-")
	if !ok || bot == "txt" {
		return false
	}
	for _, known := range Bots {
		if strings.EqualFold(known.Name, bot) && known.Purpose == PurposeSearch {
			return true
		}
	}
	return false
}

// indexingDirectiveReasons reports robots directives that block citation even
// when the scored noindex check stays green: "none", and noindex or nosnippet
// aimed at one search crawler instead of every crawler.
func indexingDirectiveReasons(sig *PageSignals) []AccessReason {
	meta := ParseMetaRobots(sig.RawHTML)
	header := ParseXRobotsTag(sig.Headers.Values("X-Robots-Tag"))
	var reasons []AccessReason
	if rulesHave(meta, pageCrawler, robotsNoneToken) || rulesHave(header, pageCrawler, robotsNoneToken) {
		reasons = append(reasons, AccessReason{ID: "robots_none", Detail: "robots none means noindex and nofollow"})
	}
	for _, bot := range Bots {
		if bot.Purpose != PurposeSearch {
			continue
		}
		if reason, ok := searchBotDirective(meta, header, bot); ok {
			reasons = append(reasons, reason)
		}
	}
	return reasons
}

func searchBotDirective(meta, header XRobotsRules, bot Bot) (AccessReason, bool) {
	agent := strings.ToLower(bot.Name)
	id := "robots-" + agent
	switch {
	case rulesHave(meta, agent, "noindex"), rulesHave(header, agent, "noindex"), rulesHave(meta, agent, robotsNoneToken), rulesHave(header, agent, robotsNoneToken):
		return AccessReason{ID: id, Detail: bot.Name + " is marked noindex"}, true
	case rulesHave(meta, agent, "nosnippet"), rulesHave(header, agent, "nosnippet"):
		return AccessReason{ID: id, Detail: bot.Name + " is marked nosnippet"}, true
	default:
		return AccessReason{}, false
	}
}

func rulesHave(rules XRobotsRules, agent, token string) bool {
	return rules[agent][token]
}
