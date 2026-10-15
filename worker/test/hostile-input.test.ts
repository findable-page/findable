import { expect, test } from "bun:test";
import { analyzeHtml } from "../src/index";

const SIZE = 1024 * 1024;
const hostile: Record<string, string> = {
  "unclosed anchors": "<a>".repeat(SIZE / 3),
  "unclosed list items": `<ul>${"<li>x".repeat(SIZE / 5)}</ul>`,
  "meta without end": "<meta ".repeat(SIZE / 6),
  "meta with late end": `${"<meta ".repeat(SIZE / 6)}>`,
  "dangling quotes": '<a "'.repeat(SIZE / 4),
  "unclosed headings": "<h1>".repeat(SIZE / 4),
  "unclosed scripts": "<script>".repeat(SIZE / 8),
  "unclosed comments": "<!--".repeat(SIZE / 4),
  "spaced angle brackets": "< a".repeat(SIZE / 3),
  "bang without name": "<!a".repeat(SIZE / 3),
  "closing without name": "</ ".repeat(SIZE / 3),
  "digit after angle bracket": "<1".repeat(SIZE / 2),
  "processing instructions": "<?".repeat(SIZE / 2),
  "unterminated doctypes": "<!doctype".repeat(SIZE / 9),
  "inline styles with endless declarations": `<p style="${"color:#777;".repeat(SIZE / 11)}">x</p>`,
  "many styled paragraphs": "<p style='color:#777;background-color:#888'>x</p>".repeat(SIZE / 50),
};

for (const [name, html] of Object.entries(hostile)) {
  test(`1MB of ${name} is analysed in under a second`, () => {
    const started = performance.now();
    const scan = analyzeHtml("https://example.com/", html, new Headers({ "content-type": "text/html" }));
    expect(performance.now() - started).toBeLessThan(1000);
    expect(scan.seo.checks.length).toBeGreaterThan(0);
  });
}
