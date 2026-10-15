import { ok, type PageCtx, skip, tagsNamed } from "./extra-checks";
import type { Check } from "./index";
import { gated, isNode, items, type Node, ofType, parseSchemaDoc, present } from "./schema-checks";

const QUESTION_LEAD =
  /^(who|what|when|where|why|how|which|can|could|does|do|did|is|are|was|were|should|will|has|have)\b/;
const DEFINITION = /\b(is|are|refers to|means|stands for)\b/;
const BLOCK_TAGS = ["h1", "h2", "h3", "h4", "h5", "h6", "p", "li"];

interface Block {
  tag: string;
  text: string;
  words: number;
}

const wordCount = (text: string): number => text.split(/\s+/).filter(Boolean).length;
const isHeading = (b: Block): boolean => /^h[1-6]$/.test(b.tag);
const isQuestion = (b: Block): boolean => {
  const lower = b.text.toLowerCase();
  return isHeading(b) && (lower.endsWith("?") || QUESTION_LEAD.test(lower));
};

function contentBlocks(c: PageCtx): Block[] {
  return BLOCK_TAGS.flatMap((tag) =>
    tagsNamed(c, tag).map((t) => {
      const text = t.text.split(/\s+/).filter(Boolean).join(" ");
      return { pos: t.pos, tag, text, words: wordCount(text) };
    }),
  )
    .sort((a, b) => a.pos - b.pos)
    .map(({ tag, text, words }) => ({ tag, text, words }));
}

function questionAnswerChecks(blocks: Block[]): Check[] {
  let questions = 0;
  let followed = 0;
  let ideal = 0;
  for (const [i, b] of blocks.entries()) {
    if (!isQuestion(b)) continue;
    questions++;
    const next = blocks[i + 1];
    if (next?.tag === "p") {
      followed++;
      if (next.words >= 20 && next.words <= 80) ideal++;
    }
  }
  if (questions === 0) {
    return [
      skip("aeo_answer_after_question", "No question headings to check"),
      skip("aeo_answer_length_ideal", "No question headings to check"),
    ];
  }
  return [
    ok(
      "aeo_answer_after_question",
      `${followed} of ${questions} question headings are followed directly by a paragraph`,
      followed === questions,
    ),
    followed === 0
      ? skip("aeo_answer_length_ideal", "No answer paragraphs to measure")
      : ok(
          "aeo_answer_length_ideal",
          `${ideal} of ${followed} answers after question headings are 20-80 words`,
          ideal * 2 >= followed,
        ),
  ];
}

function schemaDepthChecks(c: PageCtx): Check[] {
  const doc = parseSchemaDoc(c);
  const faqs = ofType(doc, "faqpage");
  const questionCount = faqs.reduce<number>((sum, faq) => sum + items(faq.mainEntity).filter(isNode).length, 0);
  const speakable: Node[] = doc.nodes.flatMap((node) =>
    items(node.speakable).map((spec) => (isNode(spec) ? spec : {})),
  );
  return [
    faqs.length === 0
      ? skip("aeo_faq_count", "No FAQPage schema to count")
      : ok("aeo_faq_count", `FAQPage lists ${questionCount} question(s) (3 or more recommended)`, questionCount >= 3),
    gated(
      ofType(doc, "howto"),
      "aeo_howto_valid",
      "HowTo schema has a name and 2+ steps with text",
      "No HowTo schema to validate",
      (n) => {
        const steps = items(n.step);
        if (!present(n.name) || steps.length < 2) return false;
        return steps.every((step) =>
          isNode(step) ? present(step.text) || present(step.name) || present(step.itemListElement) : present(step),
        );
      },
    ),
    gated(
      speakable,
      "aeo_speakable_valid",
      "Speakable markup names a cssSelector or xpath",
      "No speakable markup to validate",
      (n) => present(n.cssSelector) || present(n.xpath),
    ),
  ];
}

function passageShapeChecks(c: PageCtx, blocks: Block[]): Check[] {
  let listAfterHeading = false;
  let paragraphs = 0;
  let wall = 0;
  let listItems = 0;
  let itemWords = 0;
  let sections = 0;
  for (const [i, b] of blocks.entries()) {
    if (b.tag === "p") {
      paragraphs++;
      if (b.words > 150) wall++;
    } else if (b.tag === "li") {
      listItems++;
      itemWords += b.words;
    } else if (b.tag === "h2" || b.tag === "h3" || b.tag === "h4") {
      sections++;
    }
    if (isHeading(b) && blocks[i + 1]?.tag === "li") listAfterHeading = true;
  }
  const words = wordCount(c.text);
  const perSection = Math.floor(words / Math.max(sections, 1));
  return [
    ok("aeo_list_after_heading", "A heading is followed directly by a list", listAfterHeading),
    listItems === 0
      ? skip("aeo_list_item_length", "No list items to measure")
      : ok(
          "aeo_list_item_length",
          `List items average ${Math.floor(itemWords / listItems)} words (25 or fewer is scannable)`,
          itemWords <= 25 * listItems,
        ),
    paragraphs === 0
      ? skip("aeo_paragraph_length", "No paragraphs to measure")
      : ok("aeo_paragraph_length", `${wall} paragraph(s) longer than 150 words`, wall === 0),
    ok(
      "aeo_section_length",
      `${perSection} words per section (350 or fewer between headings is easy to cite)`,
      words <= 350 * Math.max(sections, 1),
    ),
  ];
}

function navigabilityChecks(c: PageCtx, blocks: Block[]): Check[] {
  const seen = new Map<string, number>();
  let headings = 0;
  let duplicates = 0;
  let long = 0;
  for (const b of blocks) {
    if (!isHeading(b) || b.text === "") continue;
    headings++;
    const key = b.text.toLowerCase();
    const count = (seen.get(key) ?? 0) + 1;
    seen.set(key, count);
    if (count === 2) duplicates++;
    if ([...b.text].length > 100) long++;
  }
  const subheads = [...tagsNamed(c, "h2"), ...tagsNamed(c, "h3")];
  const anchored = subheads.filter((t) => (t.attrs.id ?? "").trim() !== "").length;
  const toc = tagsNamed(c, "a").filter((a) => {
    const href = (a.attrs.href ?? "").trim();
    return href.startsWith("#") && href.length > 1;
  }).length;
  const words = wordCount(c.text);
  return [
    headings < 2
      ? skip("aeo_unique_headings", "Fewer than 2 headings to compare")
      : ok("aeo_unique_headings", `${duplicates} duplicated heading text(s)`, duplicates === 0),
    headings === 0
      ? skip("aeo_heading_length", "No headings to measure")
      : ok("aeo_heading_length", `${long} heading(s) longer than 100 characters`, long === 0),
    subheads.length === 0
      ? skip("aeo_heading_anchors", "No h2/h3 headings to anchor")
      : ok(
          "aeo_heading_anchors",
          `${anchored} of ${subheads.length} h2/h3 headings have an id for deep links`,
          anchored * 2 >= subheads.length,
        ),
    words < 800
      ? skip("aeo_toc_links", "Page is short enough to skip a table of contents")
      : ok("aeo_toc_links", `${toc} in-page anchor links (3 or more for long pages)`, toc >= 3),
  ];
}

function answerPhrasingChecks(blocks: Block[]): Check[] {
  let pairs = 0;
  let defined = false;
  for (let i = 1; i < blocks.length; i++) {
    const block = blocks[i];
    const previous = blocks[i - 1];
    if (!block || !previous || block.tag !== "p" || !isHeading(previous)) continue;
    pairs++;
    const sentence = block.text.toLowerCase().split(/[.!?]/)[0] ?? "";
    if (wordCount(sentence) <= 40 && DEFINITION.test(sentence)) defined = true;
  }
  return [
    pairs === 0
      ? skip("aeo_definition_lead", "No heading-and-paragraph sections to check")
      : ok("aeo_definition_lead", "A section opens with a definition-style sentence", defined),
  ];
}

function mediaStructureChecks(c: PageCtx): Check[] {
  const pair = (id: string, noun: string, parent: string, child: string, label: string): Check => {
    const parents = tagsNamed(c, parent).length;
    const children = tagsNamed(c, child).length;
    return parents === 0
      ? skip(id, `No ${noun} to check`)
      : ok(id, label.replace("{c}", String(children)).replace("{p}", String(parents)), children >= parents);
  };
  return [
    pair("aeo_table_captions", "tables", "table", "caption", "{c} caption(s) for {p} table(s)"),
    pair("aeo_figure_captions", "figures", "figure", "figcaption", "{c} figcaption(s) for {p} figure(s)"),
    pair(
      "aeo_details_summary",
      "details blocks",
      "details",
      "summary",
      "{c} summary element(s) for {p} details block(s)",
    ),
  ];
}

export function aeoStructureChecks(c: PageCtx): Check[] {
  const blocks = contentBlocks(c);
  return [
    ...questionAnswerChecks(blocks),
    ...schemaDepthChecks(c),
    ...passageShapeChecks(c, blocks),
    ...navigabilityChecks(c, blocks),
    ...answerPhrasingChecks(blocks),
    ...mediaStructureChecks(c),
  ];
}
