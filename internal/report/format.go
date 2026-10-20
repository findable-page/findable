package report

import (
	"encoding/json"
	"fmt"

	"github.com/findable-page/findable/internal/audit"
)

type Format string

const (
	FormatCSV   Format = "csv"
	FormatJSON  Format = "json"
	FormatSARIF Format = "sarif"
)

func ParseFormat(name string) (Format, error) {
	switch f := Format(name); f {
	case FormatCSV, FormatJSON, FormatSARIF:
		return f, nil
	default:
		return "", fmt.Errorf("unknown format %q (want csv, json or sarif)", name)
	}
}

func (f Format) Ext() string { return "." + string(f) }

func PageReport(f Format, host string, p audit.Page) string {
	switch f {
	case FormatJSON:
		return encodeJSON(p.Scan)
	case FormatSARIF:
		return SARIF([]audit.Page{p})
	default:
		return CSV(host, p.Scan)
	}
}

func SiteReport(f Format, host string, pages []audit.Page) string {
	switch f {
	case FormatJSON:
		return encodeJSON(pages)
	case FormatSARIF:
		return SARIF(pages)
	default:
		return CrawlCSV(host, pages)
	}
}

func encodeJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b) + "\n"
}
