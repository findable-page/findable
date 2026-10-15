package server

import (
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"regexp"
	"strings"
)

//go:embed web/index.html
var localWebUIHTML string

//go:embed web/i18n/*.json
var checkTextFS embed.FS

var (
	inlineScripts = regexp.MustCompile(`(?s)<script(?:\s[^>]*)?>(.*?)</script>`)
	inlineStyles  = regexp.MustCompile(`(?s)<style(?:\s[^>]*)?>(.*?)</style>`)
	uiPolicy      = strings.Join([]string{
		"default-src 'none'",
		"script-src " + hashSources(inlineScripts),
		"style-src 'unsafe-inline'",
		"style-src-elem " + hashSources(inlineStyles),
		"style-src-attr 'unsafe-inline'",
		"connect-src 'self'",
		"img-src 'self' data:",
		"base-uri 'none'",
		"form-action 'self'",
		"frame-ancestors 'none'",
	}, "; ")
)

func hashSources(block *regexp.Regexp) string {
	blocks := block.FindAllStringSubmatch(localWebUIHTML, -1)
	sources := make([]string, 0, len(blocks))
	for _, m := range blocks {
		sum := sha256.Sum256([]byte(m[1]))
		sources = append(sources, "'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'")
	}
	return strings.Join(sources, " ")
}
