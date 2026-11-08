package report

import (
	"fmt"

	"github.com/findable-page/findable/internal/audit"
)

// ParseSeverity parses -min-severity / gate severity names.
func ParseSeverity(name string) (audit.Severity, error) {
	switch audit.Severity(name) {
	case audit.SeverityCritical, audit.SeverityHigh, audit.SeverityMedium, audit.SeverityLow, audit.SeverityInfo:
		return audit.Severity(name), nil
	default:
		return "", fmt.Errorf("unknown severity %q (want critical, high, medium, low or info)", name)
	}
}
