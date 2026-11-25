package audit

import (
	"fmt"
	"regexp"
	"strings"
)

// llmsLinkItem matches a file-list entry such as `- [Title](https://…): note`.
var llmsLinkItem = regexp.MustCompile(`^[-*]\s+\[([^\]]*)\]\(([^)]*)\)`)

// llmsWantH1 is the shared failure detail for content before the H1 title.
const llmsWantH1 = "/llms.txt must start with the H1 title (# Name)"

// LLMSTxtSpecValid reports whether /llms.txt content follows the llmstxt.org
// v2 structure: the file starts with a single H1 title, an optional blockquote
// summary immediately after it, and H2 sections whose lists are Markdown
// links with absolute http(s) URLs. It returns the first violation as a
// human-readable detail; ok is true when the file conforms.
func LLMSTxtSpecValid(llms string) (ok bool, detail string) {
	var p llmsSpec
	lines := strings.Split(strings.TrimPrefix(llms, "\ufeff"), "\n")
	for i, raw := range lines {
		if d := p.lineIssue(i+1, strings.TrimSpace(raw)); d != "" {
			return false, d
		}
	}
	if !p.sawH1 {
		return false, llmsWantH1
	}
	if d := p.closeSection(); d != "" {
		return false, d
	}
	return true, ""
}

// llmsSpec tracks the structural state while walking the file top to bottom.
type llmsSpec struct {
	sawH1, sawSection, inFence bool
	sectionName                string
	sectionLine                int
	sectionHasLink             bool
}

func (p *llmsSpec) closeSection() string {
	if p.sawSection && !p.sectionHasLink {
		return fmt.Sprintf("line %d: H2 section %q lists no - [Title](url) links", p.sectionLine, p.sectionName)
	}
	return ""
}

// lineIssue validates one trimmed line and returns "" when it is fine.
func (p *llmsSpec) lineIssue(lineNo int, line string) string {
	if line == "" {
		return ""
	}
	if strings.HasPrefix(line, "```") {
		p.inFence = !p.inFence
		return ""
	}
	if p.inFence {
		return ""
	}
	isH1 := line == "#" || strings.HasPrefix(line, "# ")
	if !isH1 && !p.sawH1 {
		return fmt.Sprintf("line %d: %s", lineNo, llmsWantH1)
	}
	switch {
	case isH1 || strings.HasPrefix(line, "## "):
		return p.headingIssue(lineNo, line)
	case strings.HasPrefix(line, ">"):
		if p.sawSection {
			return fmt.Sprintf("line %d: the > summary must sit immediately after the H1, before any ## section", lineNo)
		}
	case isListItem(line):
		if p.sawSection {
			if d := listItemIssue(lineNo, line); d != "" {
				return d
			}
			p.sectionHasLink = true
		}
	}
	return ""
}

func isListItem(line string) bool {
	return strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ")
}

// headingIssue handles the H1 title and H2 section headings.
func (p *llmsSpec) headingIssue(lineNo int, line string) string {
	if strings.HasPrefix(line, "## ") {
		if d := p.closeSection(); d != "" {
			return d
		}
		p.sawSection = true
		p.sectionName = strings.TrimSpace(strings.TrimPrefix(line, "##"))
		p.sectionLine = lineNo
		p.sectionHasLink = false
		return ""
	}
	if p.sawH1 {
		return fmt.Sprintf("line %d: only one H1 title is allowed", lineNo)
	}
	if strings.TrimSpace(strings.TrimPrefix(line, "#")) == "" {
		return fmt.Sprintf("line %d: the H1 title must name the site", lineNo)
	}
	p.sawH1 = true
	return ""
}

// listItemIssue checks one file-list entry: a Markdown link with a title and
// an absolute http(s) URL.
func listItemIssue(lineNo int, line string) string {
	m := llmsLinkItem.FindStringSubmatch(line)
	if m == nil || strings.TrimSpace(m[1]) == "" || strings.TrimSpace(m[2]) == "" {
		return fmt.Sprintf("line %d: list items in H2 sections must be links like - [Title](https://example.com)", lineNo)
	}
	if !strings.HasPrefix(m[2], "http://") && !strings.HasPrefix(m[2], "https://") {
		return fmt.Sprintf("line %d: link %q must be an absolute http(s) URL", lineNo, m[2])
	}
	return ""
}
