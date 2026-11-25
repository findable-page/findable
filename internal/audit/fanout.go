package audit

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// FanoutIntent is one predictable sub-query an answer engine fans out from a topic.
type FanoutIntent struct {
	ID      string `json:"id"`
	Query   string `json:"query"`
	Covered bool   `json:"covered"`
	Where   string `json:"where,omitempty"`
}

// FanoutReport is the intent coverage matrix for one page.
type FanoutReport struct {
	Topic   string         `json:"topic"`
	Covered int            `json:"covered"`
	Total   int            `json:"total"`
	Intents []FanoutIntent `json:"intents"`
	PageURL string         `json:"url"`
}

type fanoutSpec struct {
	id      string
	query   string
	needles []string
}

type fanoutSpot struct {
	where string
	text  string
}

// QueryFanout builds a fixed sub-intent cluster for topic and marks which
// intents the page's headings, paragraphs or tables address.
func QueryFanout(pageURL, rawHTML, topic string) FanoutReport {
	topic = strings.TrimSpace(topic)
	sig := extractSignals(pageURL, rawHTML, nil, 0)
	spots := fanoutSpots(&sig)
	report := FanoutReport{Topic: topic, PageURL: pageURL}
	for _, spec := range fanoutSpecs(topic) {
		intent := FanoutIntent{ID: spec.id, Query: spec.query}
		for _, spot := range spots {
			if spotMatches(spot.text, spec.needles) {
				intent.Covered = true
				intent.Where = spot.where
				break
			}
		}
		if intent.Covered {
			report.Covered++
		}
		report.Intents = append(report.Intents, intent)
	}
	report.Total = len(report.Intents)
	return report
}

func fanoutSpecs(topic string) []fanoutSpec {
	if topic == "" {
		topic = "this topic"
	}
	return []fanoutSpec{
		{"definition", "what is " + topic, []string{"what is", topic + " is", "defined as"}},
		{"how_to", "how to " + topic, []string{"how to", "how do", "step by step"}},
		{"comparison", topic + " vs alternatives", []string{" vs ", "alternative", "compared with", "comparison"}},
		{"pros_cons", "pros and cons of " + topic, []string{"pros and cons", "advantages", "disadvantages"}},
		{"pricing", topic + " pricing", []string{"pricing", "price", "cost"}},
		{"examples", topic + " examples", []string{"for example", "for instance", "e.g."}},
		{"problems", "common problems with " + topic, []string{"common problem", "pitfall", "mistake"}},
		{"practices", "best practices for " + topic, []string{"best practice", "tips for"}},
	}
}

func fanoutSpots(sig *PageSignals) []fanoutSpot {
	var spots []fanoutSpot
	for _, heading := range sig.Headings {
		spots = append(spots, fanoutSpot{where: heading.Tag, text: heading.Text})
	}
	for i, paragraph := range sig.Paragraphs {
		spots = append(spots, fanoutSpot{where: fmt.Sprintf("p:%d", i+1), text: paragraph})
	}
	return append(spots, tableTextSpots(sig.RawHTML)...)
}

func tableTextSpots(raw string) []fanoutSpot {
	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return nil
	}
	var spots []fanoutSpot
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if text := strings.TrimSpace(collectText(n)); text != "" {
				spots = append(spots, fanoutSpot{where: n.Data, text: text})
			}
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return spots
}

func spotMatches(text string, needles []string) bool {
	lower := strings.ToLower(text)
	for _, needle := range needles {
		if strings.Contains(lower, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}
