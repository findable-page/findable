// Package config loads findable.json, the shared config file for gates,
// ignore lists, crawl settings and bot probes. Flags override config values.
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
)

type Config struct {
	Baseline string   `json:"baseline,omitempty"`
	Gate     Gate     `json:"gate,omitempty"`
	Ignore   []Rule   `json:"ignore,omitempty"`
	Crawl    Crawl    `json:"crawl,omitempty"`
	Bots     []string `json:"bots,omitempty"`
}

// Gate holds CI thresholds; the same fields exist as flags on audit/crawl.
type Gate struct {
	Overall       int    `json:"overall,omitempty"`
	SEO           int    `json:"seo,omitempty"`
	AEO           int    `json:"aeo,omitempty"`
	GEO           int    `json:"geo,omitempty"`
	NoNewFailures bool   `json:"noNewFailures,omitempty"`
	MaxSeverity   string `json:"maxSeverity,omitempty"`
}

// Rule ignores a check everywhere, or only on URLs matching a glob pattern
// (for example "/legacy/*").
type Rule struct {
	Check string `json:"check"`
	URL   string `json:"url,omitempty"`
}

type Crawl struct {
	Limit       int      `json:"limit,omitempty"`
	Concurrency int      `json:"concurrency,omitempty"`
	Include     []string `json:"include,omitempty"`
	Exclude     []string `json:"exclude,omitempty"`
}

// DefaultPath is the config file consulted automatically when present.
const DefaultPath = "findable.json"

// Load reads a config file. Unknown keys are an error so a typo in a gate
// never silently passes CI.
func Load(path string) (Config, error) {
	raw, err := os.ReadFile(path) //nolint:gosec // user-named config path
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var c Config
	if err := dec.Decode(&c); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return c, nil
}

// LoadDefault reads findable.json from the working directory if present;
// a missing file is not an error.
func LoadDefault() (Config, error) {
	if _, err := os.Stat(DefaultPath); err != nil {
		return Config{}, nil
	}
	return Load(DefaultPath)
}

// Ignores reports whether a failing check on a URL should be ignored by
// reports and gates. A rule with no URL pattern ignores the check anywhere.
func (c Config) Ignores(checkID, pageURL string) bool {
	for _, rule := range c.Ignore {
		if rule.Check != checkID {
			continue
		}
		if rule.URL == "" {
			return true
		}
		if matchURLPattern(rule.URL, pageURL) {
			return true
		}
	}
	return false
}

// matchURLPattern matches a glob against the URL path (and the whole URL
// when the pattern contains a scheme separator).
func matchURLPattern(pattern, pageURL string) bool {
	if len(pattern) > 0 && pattern[0] == '/' {
		if u, err := url.Parse(pageURL); err == nil {
			matched, err := path.Match(pattern, u.Path)
			return err == nil && matched
		}
	}
	matched, err := path.Match(pattern, pageURL)
	return err == nil && matched
}
