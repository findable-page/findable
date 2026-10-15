package audit

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	questionLeadRegex = regexp.MustCompile(`^(who|what|when|where|why|how|which|can|could|does|do|did|is|are|was|were|should|will|has|have)\b`)
	definitionRegex   = regexp.MustCompile(`\b(is|are|refers to|means|stands for)\b`)
	tldrPhrases       = []string{"tl;dr", "tldr", "in short", "short answer", "quick answer", "in a nutshell", "at a glance", "bottom line"}
)

type contentBlock struct {
	tag   string
	text  string
	words int
}

func (b contentBlock) isHeading() bool {
	return len(b.tag) == 2 && b.tag[0] == 'h'
}

func (b contentBlock) isQuestion() bool {
	lower := strings.ToLower(b.text)
	return b.isHeading() && (strings.HasSuffix(lower, "?") || questionLeadRegex.MatchString(lower))
}

func contentBlocks(s *PageSignals) []contentBlock {
	type positioned struct {
		pos   int
		block contentBlock
	}
	var found []positioned
	for _, name := range []string{"h1", "h2", "h3", "h4", "h5", "h6", "p", "li"} {
		for _, t := range s.Tags[name] {
			text := strings.Join(strings.Fields(t.Text), " ")
			found = append(found, positioned{t.Pos, contentBlock{tag: name, text: text, words: len(strings.Fields(text))}})
		}
	}
	sort.Slice(found, func(i, j int) bool { return found[i].pos < found[j].pos })
	out := make([]contentBlock, len(found))
	for i, f := range found {
		out[i] = f.block
	}
	return out
}

func aeoStructure(l *checkList, s *PageSignals) {
	blocks := contentBlocks(s)
	aeoQuestionAnswers(l, blocks)
	aeoSchemaDepth(l, s)
	aeoPassageShape(l, s, blocks)
	aeoNavigability(l, s, blocks)
	aeoAnswerPhrasing(l, s, blocks)
	aeoMediaStructure(l, s)
}

func aeoQuestionAnswers(l *checkList, blocks []contentBlock) {
	questions, followed, ideal := 0, 0, 0
	for i, b := range blocks {
		if !b.isQuestion() {
			continue
		}
		questions++
		if i+1 < len(blocks) && blocks[i+1].tag == "p" {
			followed++
			if w := blocks[i+1].words; w >= 20 && w <= 80 {
				ideal++
			}
		}
	}
	l.add("aeo_question_coverage", fmt.Sprintf("%d question-style headings (3 or more recommended)", questions), questions >= 3)
	if questions == 0 {
		l.skip("aeo_answer_after_question", "No question headings to check")
		l.skip("aeo_answer_length_ideal", "No question headings to check")
		return
	}
	l.add("aeo_answer_after_question", fmt.Sprintf("%d of %d question headings are followed directly by a paragraph", followed, questions), followed == questions)
	if followed == 0 {
		l.skip("aeo_answer_length_ideal", "No answer paragraphs to measure")
		return
	}
	l.add("aeo_answer_length_ideal", fmt.Sprintf("%d of %d answers after question headings are 20-80 words", ideal, followed), ideal*2 >= followed)
}

func aeoSchemaDepth(l *checkList, s *PageSignals) {
	doc := parseSchemaDoc(s)

	faqs := doc.ofType("faqpage")
	questionCount := 0
	for _, faq := range faqs {
		for _, q := range jsonItems(faq["mainEntity"]) {
			if _, ok := q.(map[string]any); ok {
				questionCount++
			}
		}
	}
	skipUnless(l, len(faqs) > 0, "aeo_faq_count", fmt.Sprintf("FAQPage lists %d question(s) (3 or more recommended)", questionCount), "No FAQPage schema to count", questionCount >= 3)

	howtos := doc.ofType("howto")
	skipUnless(l, len(howtos) > 0, "aeo_howto_valid", "HowTo schema has a name and 2+ steps with text", "No HowTo schema to validate", allNodes(howtos, func(n map[string]any) bool {
		steps := jsonItems(n["step"])
		if !jsonPresent(n["name"]) || len(steps) < 2 {
			return false
		}
		for _, step := range steps {
			if m, ok := step.(map[string]any); ok {
				if !jsonPresent(m["text"]) && !jsonPresent(m["name"]) && !jsonPresent(m["itemListElement"]) {
					return false
				}
			} else if !jsonPresent(step) {
				return false
			}
		}
		return true
	}))

	var speakable []map[string]any
	for _, node := range doc.nodes {
		for _, spec := range jsonItems(node["speakable"]) {
			if m, ok := spec.(map[string]any); ok {
				speakable = append(speakable, m)
			} else {
				speakable = append(speakable, map[string]any{})
			}
		}
	}
	skipUnless(l, len(speakable) > 0, "aeo_speakable_valid", "Speakable markup names a cssSelector or xpath", "No speakable markup to validate", allNodes(speakable, func(n map[string]any) bool {
		return jsonPresent(n["cssSelector"]) || jsonPresent(n["xpath"])
	}))
}

func aeoPassageShape(l *checkList, s *PageSignals, blocks []contentBlock) {
	listAfterHeading := false
	paragraphs, wall := 0, 0
	items, itemWords := 0, 0
	for i, b := range blocks {
		switch b.tag {
		case "p":
			paragraphs++
			if b.words > 150 {
				wall++
			}
		case "li":
			items++
			itemWords += b.words
		}
		if b.isHeading() && i+1 < len(blocks) && blocks[i+1].tag == "li" {
			listAfterHeading = true
		}
	}
	l.add("aeo_list_after_heading", "A heading is followed directly by a list", listAfterHeading)
	skipUnless(l, items > 0, "aeo_list_item_length", fmt.Sprintf("List items average %d words (25 or fewer is scannable)", itemWords/maxInt(items, 1)), "No list items to measure", itemWords <= 25*items)
	skipUnless(l, paragraphs > 0, "aeo_paragraph_length", fmt.Sprintf("%d paragraph(s) longer than 150 words", wall), "No paragraphs to measure", wall == 0)

	sections := 0
	for _, b := range blocks {
		if b.tag == "h2" || b.tag == "h3" || b.tag == "h4" {
			sections++
		}
	}
	words := len(s.Words)
	l.add("aeo_section_length", fmt.Sprintf("%d words per section (350 or fewer between headings is easy to cite)", words/maxInt(sections, 1)), words <= 350*maxInt(sections, 1))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func aeoNavigability(l *checkList, s *PageSignals, blocks []contentBlock) {
	headings := 0
	seen := map[string]int{}
	duplicates, long := 0, 0
	for _, b := range blocks {
		if !b.isHeading() || b.text == "" {
			continue
		}
		headings++
		key := strings.ToLower(b.text)
		seen[key]++
		if seen[key] == 2 {
			duplicates++
		}
		if len([]rune(b.text)) > 100 {
			long++
		}
	}
	skipUnless(l, headings >= 2, "aeo_unique_headings", fmt.Sprintf("%d duplicated heading text(s)", duplicates), "Fewer than 2 headings to compare", duplicates == 0)
	skipUnless(l, headings > 0, "aeo_heading_length", fmt.Sprintf("%d heading(s) longer than 100 characters", long), "No headings to measure", long == 0)

	anchored, subheads := 0, 0
	for _, name := range []string{"h2", "h3"} {
		for _, t := range s.Tags[name] {
			subheads++
			if strings.TrimSpace(t.Attrs["id"]) != "" {
				anchored++
			}
		}
	}
	skipUnless(l, subheads > 0, "aeo_heading_anchors", fmt.Sprintf("%d of %d h2/h3 headings have an id for deep links", anchored, subheads), "No h2/h3 headings to anchor", anchored*2 >= subheads)

	toc := 0
	for _, a := range s.Tags["a"] {
		if href := strings.TrimSpace(a.Attrs["href"]); strings.HasPrefix(href, "#") && len(href) > 1 {
			toc++
		}
	}
	skipUnless(l, len(s.Words) >= 800, "aeo_toc_links", fmt.Sprintf("%d in-page anchor links (3 or more for long pages)", toc), "Page is short enough to skip a table of contents", toc >= 3)
}

func aeoAnswerPhrasing(l *checkList, s *PageSignals, blocks []contentBlock) {
	lower := strings.ToLower(s.TextContent)
	tldr := false
	for _, phrase := range tldrPhrases {
		tldr = tldr || strings.Contains(lower, phrase)
	}
	l.add("aeo_tldr", "TL;DR or short-answer block present", tldr)

	pairs, defined := 0, false
	for i := 1; i < len(blocks); i++ {
		if blocks[i].tag != "p" || !blocks[i-1].isHeading() {
			continue
		}
		pairs++
		sentence := strings.ToLower(blocks[i].text)
		if cut := strings.IndexAny(sentence, ".!?"); cut >= 0 {
			sentence = sentence[:cut]
		}
		if len(strings.Fields(sentence)) <= 40 && definitionRegex.MatchString(sentence) {
			defined = true
		}
	}
	skipUnless(l, pairs > 0, "aeo_definition_lead", "A section opens with a definition-style sentence", "No heading-and-paragraph sections to check", defined)
}

func aeoMediaStructure(l *checkList, s *PageSignals) {
	tables, captions := len(s.Tags["table"]), len(s.Tags["caption"])
	skipUnless(l, tables > 0, "aeo_table_captions", fmt.Sprintf("%d caption(s) for %d table(s)", captions, tables), "No tables to caption", captions >= tables)
	figures, figcaptions := len(s.Tags["figure"]), len(s.Tags["figcaption"])
	skipUnless(l, figures > 0, "aeo_figure_captions", fmt.Sprintf("%d figcaption(s) for %d figure(s)", figcaptions, figures), "No figures to caption", figcaptions >= figures)
	details, summaries := len(s.Tags["details"]), len(s.Tags["summary"])
	skipUnless(l, details > 0, "aeo_details_summary", fmt.Sprintf("%d summary element(s) for %d details block(s)", summaries, details), "No details blocks to check", summaries >= details)
}
