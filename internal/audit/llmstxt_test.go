package audit

import (
	"strings"
	"testing"
)

const llmsTxtValid = `# Example Co

> Example Co sells hand-made ceramic mugs and ships worldwide.

## Pages

- [Shop](https://example.com/shop): every mug we sell, with sizes and prices
- [Shipping](https://example.com/shipping): countries, costs and delivery times

## Optional

- [Full text](https://example.com/llms-full.txt)
`

func TestLLMSTxtSpecValidAcceptsConformingFiles(t *testing.T) {
	files := []string{
		llmsTxtValid,
		"# Title\n",
		"# Title\n\n> Summary line\n",
		"# Title\n\nSome prose before the sections.\n\nStill prose.\n\n## Docs\n\n- [Docs](http://example.com/docs)\n",
		"# Title\n\n## Docs\n\n* [Docs](https://example.com/docs): star bullets are fine\n",
		"\ufeff# Title\n\n> BOM tolerated\n",
		"# Title\n\n## Docs\n\n```\n- not a link, but fenced\n```\n\n- [Docs](https://example.com/docs)\n",
	}
	for _, f := range files {
		if ok, detail := LLMSTxtSpecValid(f); !ok {
			t.Errorf("LLMSTxtSpecValid must accept %q, got %q", f, detail)
		}
	}
}

const mustStartWithH1 = "must start with the H1"

func TestLLMSTxtSpecValidRejectsMalformedFiles(t *testing.T) {
	cases := []struct {
		name, file, wantDetail string
	}{
		{"empty file", "", mustStartWithH1},
		{"no h1", "## Docs\n\n- [Docs](https://example.com)\n", mustStartWithH1},
		{"prose before h1", "Welcome!\n\n# Title\n", mustStartWithH1},
		{"empty h1 title", "#\n\n## Docs\n", "H1 title must name"},
		{"second h1", llmsTxtValid + "\n# Another\n", "only one H1"},
		{"summary after section", "# Title\n\n## Docs\n\n- [Docs](https://example.com)\n\n> Late summary\n", "before any ## section"},
		{"section without links", "# Title\n\n## Docs\n\nJust prose in here.\n", `section "Docs" lists no`},
		{"plain list item", "# Title\n\n## Docs\n\n- Not a link\n", "must be links like"},
		{"empty link target", "# Title\n\n## Docs\n\n- [Docs]()\n", "must be links like"},
		{"relative link", "# Title\n\n## Docs\n\n- [Docs](/docs)\n", "absolute http(s) URL"},
		{"code fence unterminated no h1", "```markdown\n# Title\n", mustStartWithH1},
	}
	for _, c := range cases {
		ok, detail := LLMSTxtSpecValid(c.file)
		if ok {
			t.Errorf("%s: expected rejection", c.name)
			continue
		}
		if !strings.Contains(detail, c.wantDetail) {
			t.Errorf("%s: detail %q must contain %q", c.name, detail, c.wantDetail)
		}
	}
}
