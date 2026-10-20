package report

import (
	"strings"

	"github.com/findable-page/findable/internal/audit"
)

const (
	sarifSchema  = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/schemas/sarif-schema-2.1.0.json"
	sarifVersion = "2.1.0"
	sarifLevel   = "warning"
	toolName     = "findable"
	toolInfoURI  = "https://github.com/findable-page/findable"
)

type sarifText struct {
	Text string `json:"text"`
}

type sarifRule struct {
	ID               string    `json:"id"`
	ShortDescription sarifText `json:"shortDescription"`
}

type sarifArtifact struct {
	URI string `json:"uri"`
}

type sarifLocation struct {
	PhysicalLocation struct {
		ArtifactLocation sarifArtifact `json:"artifactLocation"`
	} `json:"physicalLocation"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifText       `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	InformationURI string      `json:"informationUri"`
	Rules          []sarifRule `json:"rules"`
}

type sarifRun struct {
	Tool struct {
		Driver sarifDriver `json:"driver"`
	} `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifDocument struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

func SARIF(pages []audit.Page) string {
	run := sarifRun{Results: []sarifResult{}}
	run.Tool.Driver = sarifDriver{Name: toolName, InformationURI: toolInfoURI, Rules: []sarifRule{}}
	seen := map[string]bool{}
	for _, p := range pages {
		for _, c := range p.Scan.Reported() {
			for _, k := range c.Cat.Checks {
				if audit.Status(k) != audit.StatusFail {
					continue
				}
				id := strings.ToLower(c.Name) + "/" + k.ID
				if !seen[id] {
					seen[id] = true
					run.Tool.Driver.Rules = append(run.Tool.Driver.Rules, sarifRule{ID: id, ShortDescription: sarifText{Text: k.ID}})
				}
				run.Results = append(run.Results, failureResult(id, k.Label, p.URL))
			}
		}
	}
	return encodeJSON(sarifDocument{Schema: sarifSchema, Version: sarifVersion, Runs: []sarifRun{run}})
}

func failureResult(ruleID, label, uri string) sarifResult {
	var loc sarifLocation
	loc.PhysicalLocation.ArtifactLocation.URI = uri
	return sarifResult{RuleID: ruleID, Level: sarifLevel, Message: sarifText{Text: label}, Locations: []sarifLocation{loc}}
}
