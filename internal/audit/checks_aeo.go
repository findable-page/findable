package audit

import (
	"fmt"
	"strings"
)

func evalAEO(s *PageSignals) []Check {
	return runChecks(s,
		aeoStructuredData,
		aeoAnswerShape,
		aeoFormatting,
		aeoReadability,
		aeoSemantics,
	)
}

func aeoStructuredData(l *checkList, s *PageSignals) {
	l.add("aeo_structured_data", "Structured data found for AI extraction", len(s.JSONLDs) > 0)
	l.add("aeo_faq", "FAQ schema or Q&A section detected", hasSchemaType(s.SchemaTypes, "FAQPage"))
	l.add("aeo_speakable", "Speakable markup found for voice assistants", hasSpeakable(s.JSONLDs))
}

func hasSpeakable(nodes []map[string]any) bool {
	for _, node := range nodes {
		if _, ok := node["speakable"]; ok {
			return true
		}
	}
	return false
}

func aeoAnswerShape(l *checkList, s *PageSignals) {
	l.add("aeo_direct_answers", "Concise answer paragraphs (30-80 words) found", hasParagraphOfLength(s.Paragraphs, 30, 80))
	l.add("aeo_question_headings", "Question-style headings found (What/How/Why/?)", hasQuestionHeading(s.Headings))
}

func hasQuestionHeading(headings []Heading) bool {
	for _, h := range headings {
		t := strings.ToLower(h.Text)
		if strings.HasPrefix(t, "what") || strings.HasPrefix(t, "how") || strings.HasPrefix(t, "why") || strings.HasSuffix(t, "?") {
			return true
		}
	}
	return false
}

func aeoFormatting(l *checkList, s *PageSignals) {
	l.add("aeo_definition_lists", "Definition lists (<dl>) present for terminology", s.DLCount > 0)
	l.add("aeo_numbered_steps", "Numbered steps (<ol>) present for instructions", s.OLCount > 0)
	l.add("aeo_tables", "Data tables present for comparison", s.TableCount > 0)
	l.add("aeo_lists", "Lists (<ul>/<ol>) present for scannability", s.ULCount > 0 || s.OLCount > 0)
	l.add("aeo_code_blocks", "Code blocks (<code>/<pre>) present for technical clarity", s.CodeCount > 0)
	l.add("aeo_collapsible", "<details>/<summary> elements present", s.DetailsCount > 0)
}

func aeoReadability(l *checkList, s *PageSignals) {
	lowerText := strings.ToLower(s.TextContent)
	pronouns := strings.Count(lowerText, " you ") + strings.Count(lowerText, " your ") + strings.Count(lowerText, " we ")
	l.add("aeo_conversational", "Conversational tone with user pronouns ('you', 'your', 'we')", pronouns >= 3)

	sentences := strings.Split(s.TextContent, ".")
	avgSentenceLen := float64(len(s.Words)) / float64(len(sentences))
	l.add("aeo_reading_level", fmt.Sprintf("Average sentence length: %.1f words (<20 is AI-friendly)", avgSentenceLen), avgSentenceLen <= 22.0)
	l.add("aeo_short_sentences", "Concise sentence structures dominant", avgSentenceLen <= 20.0)

	l.add("aeo_summary", "Summary or Key Takeaways section present", lowerContainsAny(s.TextContent, "key takeaway", "summary", "conclusion"))
}

func aeoSemantics(l *checkList, s *PageSignals) {
	cleanURL := s.URL != nil && !strings.Contains(s.URL.Path, "_") && strings.ToLower(s.URL.Path) == s.URL.Path
	l.add("aeo_clean_url", "Clean lowercase URL structure", cleanURL)

	semanticTotal := 0
	for _, count := range s.SemanticCounts {
		semanticTotal += count
	}
	l.add("aeo_semantic_html", fmt.Sprintf("%d semantic HTML5 elements found", semanticTotal), semanticTotal >= 4)
	l.add("aeo_ai_robots", "No AI-blocking directives detected", !lowerContainsAny(s.RobotsMeta, "noindex"))
}
