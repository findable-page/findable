package audit

import "strings"

// Bot purposes: how the crawler's operator uses the fetched content.
const (
	PurposeSearch    = "search"     // indexes content for AI answers / search
	PurposeUserFetch = "user-fetch" // fetches when a user asks, on demand
	PurposeTraining  = "training"   // trains models; no effect on AI search
)

// Bot describes one crawler: its user agent, what it is for and the vendor
// doc that explains it. One table drives the site checks, the meta registry
// and the live probe.
type Bot struct {
	Name        string
	UA          string
	Purpose     string
	Doc         string
	Consequence string // what happens if you block it
}

const (
	openAIBotsDoc = "https://developers.openai.com/api/docs/bots"
	claudeBotsDoc = "https://support.claude.com/en/articles/8896518"
)

// Bots is the default probe list, ordered: search, user-fetch, training.
var Bots = []Bot{
	{
		"OAI-SearchBot", "OAI-SearchBot/1.0; +https://openai.com/bot", PurposeSearch,
		openAIBotsDoc, "blocking it removes the site from ChatGPT Search",
	},
	{
		"PerplexityBot", "Mozilla/5.0 (compatible; PerplexityBot/1.0; +https://perplexity.ai/perplexitybot)", PurposeSearch,
		"https://docs.perplexity.ai/docs/perplexity-crawlers", "blocking it removes the site from Perplexity answers",
	},
	{
		"Claude-SearchBot", "Mozilla/5.0 (compatible; Claude-SearchBot/1.0; +https://www.anthropic.com)", PurposeSearch,
		claudeBotsDoc, "blocking it removes the site from Claude web search",
	},
	{
		"Googlebot", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", PurposeSearch,
		"https://developers.google.com/search/docs/crawling-indexing/overview-google-crawlers", "blocking it removes the site from Google Search and AI Overviews",
	},
	{
		"Bingbot", "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)", PurposeSearch,
		"https://www.bing.com/webmasters/help/bingbot-1751", "blocking it removes the site from Bing (a ChatGPT search provider) or lowers it",
	},

	{
		"ChatGPT-User", "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; ChatGPT-User/1.0; +https://openai.com/bot", PurposeUserFetch,
		openAIBotsDoc, "blocking it makes user-pasted links fail to load in ChatGPT",
	},
	{
		"Perplexity-User", "Mozilla/5.0 (compatible; Perplexity-User/1.0; +https://perplexity.ai/perplexitybot)", PurposeUserFetch,
		"https://docs.perplexity.ai/docs/perplexity-crawlers", "blocking it makes user-pasted links fail to load in Perplexity",
	},
	{
		"Claude-User", "Mozilla/5.0 (compatible; Claude-User/1.0; +https://www.anthropic.com)", PurposeUserFetch,
		claudeBotsDoc, "blocking it makes user-pasted links fail to load in Claude",
	},

	{
		"GPTBot", "GPTBot/1.0; +https://openai.com/gptbot", PurposeTraining,
		openAIBotsDoc, "blocking it only opts out of OpenAI model training",
	},
	{
		"ClaudeBot", "ClaudeBot/1.0; +claude-bot@anthropic.com", PurposeTraining,
		claudeBotsDoc, "blocking it only opts out of Anthropic model training",
	},
	{
		"Google-Extended", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", PurposeTraining,
		"https://developers.google.com/search/docs/crawling-indexing/google-extended", "blocking it opts out of Gemini training; it does not control AI Overviews",
	},
	{
		"Applebot-Extended", "Mozilla/5.0 (compatible; Applebot/0.1; +http://www.apple.com/go/applebot)", PurposeTraining,
		"https://support.apple.com/en-us/119829", "blocking it opts out of Apple model training",
	},
	{
		"meta-externalagent", "meta-externalagent/1.1 (+https://developers.facebook.com/docs/sharing/webmasters/crawler)", PurposeTraining,
		"https://developers.facebook.com/docs/sharing/webmasters/web-crawlers", "blocking it opts out of Meta AI training",
	},
	{
		"CCBot", "CCBot/2.0 (https://commoncrawl.org/faq/)", PurposeTraining,
		"https://commoncrawl.org/faq/", "blocking it opts out of Common Crawl data",
	},
}

// BotByName looks a bot up in the table (case-insensitive).
func BotByName(name string) (Bot, bool) {
	for _, b := range Bots {
		if strings.EqualFold(b.Name, name) {
			return b, true
		}
	}
	return Bot{}, false
}

func botsWithPurpose(purpose string) []string {
	var out []string
	for _, b := range Bots {
		if b.Purpose == purpose {
			out = append(out, b.Name)
		}
	}
	return out
}

// aiSearchCrawlers retrieve page content to cite it in AI answers
// (ChatGPT Search, Perplexity, Claude, Google Search and AI Overviews, Bing).
var aiSearchCrawlers = botsWithPurpose(PurposeSearch)

// userFetchCrawlers fetch pages only when a user pastes or requests them.
var userFetchCrawlers = botsWithPurpose(PurposeUserFetch)

// trainingCrawlers fetch page content to train AI models. Blocking them does
// not remove a site from AI search answers while the search crawlers stay
// allowed.
var trainingCrawlers = botsWithPurpose(PurposeTraining)
