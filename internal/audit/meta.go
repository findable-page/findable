package audit

import "strings"

// Severity ranks how much a failing check hurts findability. Weights drive
// the score model: critical 8, high 4, medium 2, low 1, info 0.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Weight is the scoring weight of the severity level.
func (s Severity) Weight() float64 {
	switch s {
	case SeverityCritical:
		return 8
	case SeverityHigh:
		return 4
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	}
	return 0
}

// Tier states how much evidence backs a check, and scales its weight:
// documented 1.0 (vendor docs say so), reported 0.75 (independent tests),
// heuristic 0.5 (our judgment). This is the anti-snake-oil stance: reports
// show the tier so users know what is measured truth and what is advice.
type Tier string

const (
	TierDocumented Tier = "documented"
	TierReported   Tier = "reported"
	TierHeuristic  Tier = "heuristic"
)

// Factor scales a check's weight by how strong its evidence is.
func (t Tier) Factor() float64 {
	switch t {
	case TierDocumented:
		return 1.0
	case TierReported:
		return 0.75
	}
	return 0.5
}

// Meta is the registry entry for a check ID: how severe a failure is, how
// strong the evidence is, and the one-line fix a user should act on.
type Meta struct {
	Severity Severity
	Tier     Tier
	Fix      string
	Doc      string // /learn slug on the findable site, optional
}

var registry = map[string]Meta{
	// SEO: titles, descriptions, head tags
	"title_exists":          {SeverityHigh, TierDocumented, "Add a <title> tag in <head> that names the page.", ""},
	"title_length":          {SeverityLow, TierHeuristic, "Keep the title between 30 and 60 characters.", ""},
	"title_single":          {SeverityMedium, TierDocumented, "Keep exactly one <title> tag in <head>.", ""},
	"desc_exists":           {SeverityMedium, TierDocumented, "Add a meta description that summarizes the page for search snippets.", ""},
	"desc_length":           {SeverityLow, TierHeuristic, "Keep the meta description between 120 and 160 characters.", ""},
	"desc_single":           {SeverityMedium, TierDocumented, "Keep exactly one meta description tag.", ""},
	"charset":               {SeverityMedium, TierDocumented, "Declare <meta charset=\"utf-8\"> so content is decoded correctly.", ""},
	"charset_early":         {SeverityLow, TierDocumented, "Declare the charset within the first 1024 bytes of the document.", ""},
	"lang_attr":             {SeverityMedium, TierDocumented, "Set a lang attribute on the <html> element.", ""},
	"lang_valid":            {SeverityLow, TierHeuristic, "Use a valid language tag (for example en or en-US) in the lang attribute.", ""},
	"html_doctype":          {SeverityLow, TierDocumented, "Start the document with <!DOCTYPE html> to get standards mode.", ""},
	"meta_refresh":          {SeverityHigh, TierDocumented, "Replace the meta refresh with a server-side redirect.", ""},
	"viewport":              {SeverityMedium, TierDocumented, "Add <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">.", ""},
	"viewport_device_width": {SeverityMedium, TierDocumented, "Set the viewport width to device-width.", ""},
	"viewport_scalable":     {SeverityMedium, TierDocumented, "Do not block pinch zoom (avoid user-scalable=no and maximum-scale below 2).", ""},
	"favicon":               {SeverityLow, TierDocumented, "Link a favicon from <head>.", ""},
	"url_length":            {SeverityLow, TierHeuristic, "Keep URL paths and queries under about 100 characters.", ""},
	"url_query_params":      {SeverityLow, TierHeuristic, "Keep URLs to two or fewer query parameters.", ""},
	"url_depth":             {SeverityLow, TierHeuristic, "Keep pages within five path levels of the root.", ""},
	"hreflang_valid":        {SeverityHigh, TierDocumented, "Fix invalid hreflang language codes.", ""},
	"hreflang_absolute":     {SeverityMedium, TierDocumented, "Use absolute URLs in hreflang annotations.", ""},

	// SEO: robots and canonicals
	"robots_meta":                    {SeverityCritical, TierDocumented, "Remove noindex from the robots meta tag so the page can be indexed.", ""},
	"robots_nofollow":                {SeverityHigh, TierDocumented, "Remove nofollow or none from the robots meta tag so links can be followed.", ""},
	"meta_robots_conflict":           {SeverityHigh, TierDocumented, "Remove contradictory robots meta directives such as index together with noindex.", ""},
	"meta_robots_snippet_directives": {SeverityMedium, TierDocumented, "Use max-snippet:-1, max-image-preview:large and max-video-preview:-1 in the robots meta.", ""},
	"canonical":                      {SeverityMedium, TierDocumented, "Add a <link rel=\"canonical\"> pointing at this page's preferred URL.", ""},
	"canonical_self":                 {SeverityHigh, TierDocumented, "Point the canonical at this page's own URL, not a different page.", ""},
	"canonical_absolute":             {SeverityMedium, TierDocumented, "Use an absolute URL for the canonical link.", ""},
	"og_url_matches_canonical":       {SeverityMedium, TierDocumented, "Make og:url match the canonical URL.", ""},

	// SEO: social meta
	"og_title":          {SeverityLow, TierDocumented, "Add an og:title meta tag.", ""},
	"og_desc":           {SeverityLow, TierDocumented, "Add an og:description meta tag.", ""},
	"og_image":          {SeverityLow, TierDocumented, "Add an og:image meta tag.", ""},
	"og_image_absolute": {SeverityLow, TierDocumented, "Make og:image an absolute http(s) URL.", ""},
	"og_site_name":      {SeverityLow, TierDocumented, "Add og:site_name naming the site.", ""},
	"og_type":           {SeverityLow, TierDocumented, "Set og:type (for example website or article).", ""},
	"og_url":            {SeverityLow, TierDocumented, "Add og:url with the page's canonical URL.", ""},
	"twitter_card":      {SeverityLow, TierDocumented, "Add a twitter:card meta tag such as summary_large_image.", ""},

	// SEO: security headers
	"https":           {SeverityHigh, TierDocumented, "Serve the page over HTTPS.", ""},
	"hsts":            {SeverityMedium, TierDocumented, "Send a Strict-Transport-Security header so browsers only use HTTPS.", ""},
	"x_content_type":  {SeverityLow, TierDocumented, "Send X-Content-Type-Options: nosniff.", ""},
	"x_frame_options": {SeverityLow, TierDocumented, "Send X-Frame-Options (or frame-ancestors in CSP) to prevent framing.", ""},
	"csp":             {SeverityMedium, TierDocumented, "Send a Content-Security-Policy header.", ""},
	"mixed_content":   {SeverityHigh, TierDocumented, "Load every subresource over HTTPS on an HTTPS page.", ""},

	// SEO: headings and content
	"h1_exists":          {SeverityMedium, TierHeuristic, "Use exactly one H1 that describes the page.", ""},
	"h2_exists":          {SeverityLow, TierHeuristic, "Break the content into H2 sections.", ""},
	"heading_hierarchy":  {SeverityLow, TierHeuristic, "Do not skip heading levels (go H1, H2, H3).", ""},
	"word_count":         {SeverityMedium, TierHeuristic, "Write at least 300 words of body copy.", ""},
	"text_html_ratio":    {SeverityLow, TierHeuristic, "Reduce markup bloat so visible text is at least 15% of the HTML.", ""},
	"dom_size":           {SeverityLow, TierHeuristic, "Keep the DOM under about 1500 elements.", ""},
	"geo_content_length": {SeverityMedium, TierHeuristic, "Write at least 300 words so there is something to cite.", ""},

	// SEO: links
	"internal_links":      {SeverityLow, TierHeuristic, "Link to at least two internal pages.", ""},
	"external_links":      {SeverityLow, TierHeuristic, "Cite at least one external source.", ""},
	"empty_hrefs":         {SeverityMedium, TierHeuristic, "Give every <a> a real href.", ""},
	"link_noopener":       {SeverityLow, TierDocumented, "Add rel=\"noopener\" to links that open in a new tab.", ""},
	"nofollow_internal":   {SeverityMedium, TierDocumented, "Remove rel=nofollow from internal links.", ""},
	"generic_anchor_text": {SeverityLow, TierHeuristic, "Replace generic anchors such as \"click here\" with descriptive text.", ""},
	"empty_anchor_text":   {SeverityMedium, TierHeuristic, "Give every link text, an aria-label or an image.", ""},
	"javascript_hrefs":    {SeverityMedium, TierDocumented, "Use real URL hrefs instead of javascript: so crawlers can follow links.", ""},
	"links_count":         {SeverityLow, TierHeuristic, "Keep the page under about 200 links.", ""},

	// SEO: images
	"img_alt":           {SeverityMedium, TierDocumented, "Add descriptive alt text to every informative image.", ""},
	"img_dimensions":    {SeverityLow, TierHeuristic, "Set width and height on images to prevent layout shift.", ""},
	"img_alt_quality":   {SeverityLow, TierHeuristic, "Replace filename-like alt text with a real description.", ""},
	"img_alt_length":    {SeverityLow, TierHeuristic, "Keep alt text under 125 characters.", ""},
	"img_lazy_loading":  {SeverityLow, TierHeuristic, "Add loading=\"lazy\" to below-the-fold images.", ""},
	"img_modern_format": {SeverityLow, TierHeuristic, "Serve images as WebP or AVIF.", ""},

	// SEO: structure, schema, accessibility
	"jsonld":               {SeverityMedium, TierHeuristic, "Add JSON-LD structured data describing the page.", ""},
	"jsonld_valid":         {SeverityHigh, TierDocumented, "Fix the JSON syntax so every JSON-LD block parses.", ""},
	"jsonld_context":       {SeverityMedium, TierDocumented, "Include \"@context\": \"https://schema.org\" in each JSON-LD block.", ""},
	"aria_landmarks":       {SeverityLow, TierHeuristic, "Use main and nav landmark elements.", ""},
	"duplicate_ids":        {SeverityMedium, TierDocumented, "Fix duplicate id attributes; they break anchors and ARIA references.", ""},
	"skip_link":            {SeverityMedium, TierDocumented, "Add a skip-to-content link as the first focusable element.", ""},
	"button_names":         {SeverityMedium, TierDocumented, "Give every button an accessible name.", ""},
	"table_headers":        {SeverityLow, TierDocumented, "Add <th> header cells to data tables.", ""},
	"iframe_title":         {SeverityLow, TierDocumented, "Give every iframe a title attribute.", ""},
	"color_contrast":       {SeverityMedium, TierDocumented, "Raise text contrast to at least WCAG AA (4.5:1 body text).", ""},
	"keyphrase_in_title":   {SeverityLow, TierHeuristic, "Put the focus keyphrase in the title.", ""},
	"keyphrase_in_desc":    {SeverityLow, TierHeuristic, "Put the focus keyphrase in the meta description.", ""},
	"keyphrase_in_h1":      {SeverityLow, TierHeuristic, "Put the focus keyphrase in the H1.", ""},
	"keyphrase_in_content": {SeverityLow, TierHeuristic, "Use the focus keyphrase in the body copy.", ""},

	// SEO: schema.org details
	"schema_empty_values":   {SeverityLow, TierDocumented, "Fill empty structured data values instead of shipping blank strings.", ""},
	"schema_urls_absolute":  {SeverityLow, TierDocumented, "Use absolute URLs in structured data.", ""},
	"org_name":              {SeverityLow, TierDocumented, "Add a name to Organization JSON-LD.", ""},
	"org_url":               {SeverityLow, TierDocumented, "Add a url to Organization JSON-LD.", ""},
	"org_logo":              {SeverityLow, TierDocumented, "Add a logo to Organization JSON-LD.", ""},
	"website_name_url":      {SeverityLow, TierDocumented, "Set name and url on WebSite JSON-LD.", ""},
	"searchaction_valid":    {SeverityLow, TierDocumented, "Fix the SearchAction markup (action URL with {search_term_string}).", ""},
	"breadcrumb_valid":      {SeverityMedium, TierDocumented, "Fix the BreadcrumbList markup so every level has a name and item URL.", ""},
	"faq_valid":             {SeverityMedium, TierDocumented, "Fix the FAQPage markup so every question has an acceptedAnswer.", ""},
	"article_headline":      {SeverityMedium, TierDocumented, "Set headline to the article title in Article JSON-LD.", ""},
	"article_author":        {SeverityMedium, TierDocumented, "Add an author (Person) to Article JSON-LD.", ""},
	"article_dates":         {SeverityMedium, TierDocumented, "Add datePublished (and dateModified) to Article JSON-LD.", ""},
	"article_image":         {SeverityLow, TierDocumented, "Add an image to Article JSON-LD.", ""},
	"product_offers":        {SeverityMedium, TierDocumented, "Add offers to Product JSON-LD so prices can appear in results.", ""},
	"product_price":         {SeverityMedium, TierDocumented, "Add a price inside offers in Product JSON-LD.", ""},
	"product_rating":        {SeverityLow, TierDocumented, "Add aggregateRating or review to Product JSON-LD.", ""},
	"localbusiness_address": {SeverityMedium, TierDocumented, "Add postalAddress to LocalBusiness JSON-LD.", ""},
	"localbusiness_phone":   {SeverityLow, TierDocumented, "Add telephone to LocalBusiness JSON-LD.", ""},
	"video_fields":          {SeverityLow, TierDocumented, "Add thumbnailUrl, uploadDate and contentUrl to VideoObject JSON-LD.", ""},

	// SEO: page weight
	"render_blocking_scripts": {SeverityLow, TierHeuristic, "Add defer or async to scripts in <head>.", ""},
	"script_count":            {SeverityLow, TierHeuristic, "Reduce the number of external scripts.", ""},
	"stylesheet_count":        {SeverityLow, TierHeuristic, "Reduce the number of stylesheets.", ""},
	"inline_css_size":         {SeverityLow, TierHeuristic, "Move large inline styles into a cacheable stylesheet.", ""},
	"resource_hints":          {SeverityLow, TierHeuristic, "Preconnect to third-party origins used early.", ""},
	"html_weight":             {SeverityLow, TierHeuristic, "Keep HTML under 500 KB.", ""},

	// AEO
	"aeo_structured_data":       {SeverityMedium, TierHeuristic, "Add JSON-LD structured data describing the page.", ""},
	"aeo_faq":                   {SeverityLow, TierHeuristic, "Add a visible question-and-answer section or FAQPage JSON-LD.", ""},
	"aeo_faq_count":             {SeverityInfo, TierHeuristic, "Add more visible question-and-answer pairs.", ""},
	"aeo_speakable":             {SeverityLow, TierHeuristic, "Add speakable schema marking the key passages for voice assistants.", ""},
	"aeo_speakable_valid":       {SeverityLow, TierDocumented, "Fix the speakable markup so it uses valid CSS selectors.", ""},
	"aeo_direct_answers":        {SeverityMedium, TierHeuristic, "Open key sections with a 30-80 word direct answer paragraph.", ""},
	"aeo_lead_direct_answer":    {SeverityMedium, TierHeuristic, "Open the page with a 20-80 word direct-answer paragraph within 100 words of the H1.", ""},
	"aeo_question_headings":     {SeverityLow, TierHeuristic, "Phrase some headings as the questions users actually ask.", ""},
	"aeo_answer_after_question": {SeverityLow, TierHeuristic, "Answer each question heading directly with a following paragraph.", ""},
	"aeo_answer_length_ideal":   {SeverityLow, TierHeuristic, "Keep answers after question headings between 20 and 80 words.", ""},
	"aeo_definition_lists":      {SeverityInfo, TierHeuristic, "Mark up terminology with <dl> definition lists.", ""},
	"aeo_definition_lead":       {SeverityInfo, TierHeuristic, "Start definitions with the term being defined.", ""},
	"aeo_numbered_steps":        {SeverityInfo, TierHeuristic, "Use <ol> ordered lists for step-by-step instructions.", ""},
	"aeo_tables":                {SeverityInfo, TierHeuristic, "Use tables for comparative data.", ""},
	"aeo_lists":                 {SeverityInfo, TierHeuristic, "Use lists so content is scannable.", ""},
	"aeo_list_after_heading":    {SeverityInfo, TierHeuristic, "Follow list-promising headings with an actual list.", ""},
	"aeo_list_item_length":      {SeverityInfo, TierHeuristic, "Keep list items concise.", ""},
	"aeo_code_blocks":           {SeverityInfo, TierHeuristic, "Use code blocks for technical examples.", ""},
	"aeo_collapsible":           {SeverityInfo, TierHeuristic, "Wrap optional detail in <details> elements.", ""},
	"aeo_details_summary":       {SeverityInfo, TierHeuristic, "Give every <details> a <summary> label.", ""},
	"aeo_conversational":        {SeverityLow, TierHeuristic, "Write with direct pronouns (you, your) for conversational queries.", ""},
	"aeo_reading_level":         {SeverityLow, TierHeuristic, "Shorten sentences; aim for about 20 words on average.", ""},
	"aeo_short_sentences":       {SeverityLow, TierHeuristic, "Prefer short sentences.", ""},
	"aeo_summary":               {SeverityLow, TierHeuristic, "Add a summary or key-takeaways section.", ""},
	"aeo_section_length":        {SeverityLow, TierHeuristic, "Split sections over about 350 words with subheadings.", ""},
	"aeo_heading_anchors":       {SeverityInfo, TierHeuristic, "Add id attributes to headings for deep links.", ""},
	"aeo_heading_length":        {SeverityLow, TierHeuristic, "Keep headings short (under about 80 characters).", ""},
	"aeo_paragraph_length":      {SeverityLow, TierHeuristic, "Keep paragraphs under about 150 words.", ""},
	"aeo_toc_links":             {SeverityInfo, TierHeuristic, "Add a table of contents with anchor links.", ""},
	"aeo_unique_headings":       {SeverityLow, TierHeuristic, "Avoid repeating the same heading text.", ""},
	"aeo_figure_captions":       {SeverityInfo, TierHeuristic, "Add <figcaption> to figures.", ""},
	"aeo_table_captions":        {SeverityInfo, TierHeuristic, "Add <caption> to tables.", ""},
	"aeo_howto_valid":           {SeverityMedium, TierDocumented, "Fix the HowTo JSON-LD markup.", ""},
	"aeo_clean_url":             {SeverityLow, TierHeuristic, "Use lowercase URLs without underscores.", ""},
	"aeo_semantic_html":         {SeverityLow, TierHeuristic, "Use semantic HTML5 elements such as main, article and nav.", ""},
	"aeo_ai_robots":             {SeverityCritical, TierDocumented, "Remove the noindex robots meta so AI answer engines can cite this page.", ""},

	// GEO
	"geo_statistics_density":          {SeverityLow, TierHeuristic, "Add concrete numbers and units to support claims.", ""},
	"geo_data_viz":                    {SeverityInfo, TierHeuristic, "Present data in tables or charts.", ""},
	"geo_comparison":                  {SeverityInfo, TierHeuristic, "Add comparison sections or tables.", ""},
	"geo_expert_quotes":               {SeverityLow, TierHeuristic, "Quote named experts with attribution.", ""},
	"geo_citations":                   {SeverityMedium, TierHeuristic, "Cite external sources for key claims.", ""},
	"geo_definitions":                 {SeverityLow, TierHeuristic, "Define key concepts in plain sentences.", ""},
	"geo_passage_citability":          {SeverityMedium, TierHeuristic, "Write standalone 40-120 word passages that can be quoted as answers.", ""},
	"geo_author_schema":               {SeverityMedium, TierHeuristic, "Add author Person JSON-LD with a name and sameAs links.", ""},
	"geo_author_credentials":          {SeverityLow, TierHeuristic, "Show author bylines with credentials.", ""},
	"geo_org_schema":                  {SeverityMedium, TierHeuristic, "Declare Organization JSON-LD with a consistent identity.", ""},
	"geo_knowledge_graph":             {SeverityLow, TierHeuristic, "Add sameAs links to authoritative profiles (Wikipedia, LinkedIn).", ""},
	"geo_schema_knows_about":          {SeverityLow, TierHeuristic, "List 3 or more core topics in knowsAbout on Organization or WebSite JSON-LD.", ""},
	"geo_author_credential_depth":     {SeverityLow, TierHeuristic, "Give author Person JSON-LD a jobTitle, alumniOf, hasCredential or professional sameAs link.", ""},
	"geo_trust":                       {SeverityLow, TierHeuristic, "Link About and Contact pages.", ""},
	"geo_policy_links":                {SeverityLow, TierHeuristic, "Link privacy and terms pages from the footer.", ""},
	"geo_entity_markup":               {SeverityLow, TierHeuristic, "Declare more than one schema.org entity type on the page.", ""},
	"geo_freshness":                   {SeverityLow, TierHeuristic, "Show publication or update dates.", ""},
	"geo_time_element":                {SeverityLow, TierHeuristic, "Mark dates with <time datetime> so machines can read them.", ""},
	"geo_visible_date":                {SeverityLow, TierHeuristic, "Show a visible date near the content.", ""},
	"geo_date_published":              {SeverityLow, TierDocumented, "Add datePublished to Article JSON-LD.", ""},
	"geo_date_modified":               {SeverityLow, TierDocumented, "Add dateModified to Article JSON-LD.", ""},
	"geo_modified_after_published":    {SeverityLow, TierDocumented, "Keep dateModified at or after datePublished.", ""},
	"geo_intent_clarity":              {SeverityLow, TierHeuristic, "Align the title with the H1 so the page's subject is unambiguous.", ""},
	"geo_format_variety":              {SeverityInfo, TierHeuristic, "Vary formatting: lists, tables, quotes and definitions.", ""},
	"geo_snippet_control":             {SeverityMedium, TierDocumented, "Remove nosnippet or max-snippet:0 so search and AI snippets can quote the page.", ""},
	"geo_snippet_boilerplate_control": {SeverityLow, TierDocumented, "Wrap long nav/header/footer text in data-nosnippet div/span/section elements.", ""},
	"geo_quote_attribution":           {SeverityLow, TierHeuristic, "Attribute quotes to a named source.", ""},
	"geo_source_diversity":            {SeverityInfo, TierHeuristic, "Cite several distinct domains.", ""},
	"geo_authoritative_links":         {SeverityLow, TierHeuristic, "Link authoritative sources such as standards and research.", ""},

	// site-level (robots.txt, llms.txt)
	"robots-txt":           {SeverityLow, TierDocumented, "Publish a robots.txt at the site root so crawlers can read your rules.", ""},
	"robots-page":          {SeverityCritical, TierDocumented, "Allow crawlers to fetch this page in robots.txt; a Disallow here removes it from search.", ""},
	"x_robots_tag_noindex": {SeverityCritical, TierDocumented, "Remove noindex from the X-Robots-Tag response header so the page can be indexed.", ""},
	"llms-txt":             {SeverityLow, TierHeuristic, "Optionally publish /llms.txt describing the site for AI tools; it is not a ranking signal.", ""},
}

// MetaFor returns registry metadata for a check ID. Dynamic robots-<bot>
// checks get their meta from the crawler classification; anything unknown
// falls back to medium/heuristic so new checks still score honestly.
func MetaFor(id string) Meta {
	if m, ok := registry[id]; ok {
		return m
	}
	if bot, ok := strings.CutPrefix(id, "robots-"); ok && bot != "" && bot != "txt" && bot != "page" {
		return robotsBotMeta(bot)
	}
	return Meta{Severity: SeverityMedium, Tier: TierHeuristic, Fix: "Fix the issue described in the check label."}
}

func robotsBotMeta(bot string) Meta {
	for _, name := range aiSearchCrawlers {
		if strings.EqualFold(name, bot) {
			return Meta{SeverityHigh, TierDocumented, "Allow " + name + " in robots.txt; blocking it removes the site from that engine's AI answers.", ""}
		}
	}
	for _, name := range userFetchCrawlers {
		if strings.EqualFold(name, bot) {
			return Meta{SeverityLow, TierDocumented, "Allow " + name + " in robots.txt so users can paste your links into that assistant.", ""}
		}
	}
	return Meta{SeverityInfo, TierDocumented, "Decide deliberately whether to allow " + bot + " (a training crawler); blocking it does not affect AI search citations.", ""}
}

// RegistryIDs lists all static registry IDs plus the dynamic robots-bot
// family, so callers can hash the registry for baseline compatibility.
func RegistryIDs() []string {
	ids := make([]string, 0, len(registry)+len(Bots))
	for id := range registry {
		ids = append(ids, id)
	}
	for _, b := range Bots {
		ids = append(ids, "robots-"+strings.ToLower(b.Name))
	}
	return ids
}
