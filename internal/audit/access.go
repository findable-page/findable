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
	for _, named := range s.Reported() {
		for _, check := range named.Cat.Checks {
			if gate1Failure(check) {
				reasons = append(reasons, AccessReason{ID: check.ID, Detail: check.Label})
			}
		}
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
