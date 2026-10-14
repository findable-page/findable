import { expect, test } from "bun:test";
import { robotsAllows, siteCategory } from "../src/site";

const robotsBlockingAI = `User-agent: *
Disallow: /private/
Allow: /private/open$

User-agent: GPTBot
User-agent: ClaudeBot
Disallow: /

User-agent: PerplexityBot
Disallow: /*.pdf$
`;

test("robotsAllows uses the most specific group and the longest rule", () => {
  const cases: [string, string, boolean][] = [
    ["Googlebot", "/", true],
    ["Googlebot", "/private/x", false],
    ["Googlebot", "/private/open", true],
    ["Googlebot", "/private/open/more", false],
    ["GPTBot", "/blog", false],
    ["claudebot", "/", false],
    ["PerplexityBot", "/private/x", true],
    ["PerplexityBot", "/doc.pdf", false],
    ["Google-Extended", "/private/x", false],
  ];
  for (const [agent, path, want] of cases) {
    expect([agent, path, robotsAllows(robotsBlockingAI, agent, path)]).toEqual([agent, path, want]);
  }
});

test("robotsAllows matches the product token before a version", () => {
  expect(robotsAllows("User-agent: GPTBot/1.0\nDisallow: /\n", "GPTBot", "/")).toBe(false);
});

test("robotsAllows allows everything when robots is empty or Disallow is blank", () => {
  for (const robots of ["", "User-agent: *\nDisallow:\n", "# just a comment\n"]) {
    expect(robotsAllows(robots, "GPTBot", "/any")).toBe(true);
  }
});

test("siteCategory matches the Go engine check ids and counts", () => {
  const cat = siteCategory("/blog", robotsBlockingAI, "# Example\n");
  const status = Object.fromEntries(cat.checks.map((c) => [c.id, c.pass]));
  expect(status).toEqual({
    "robots-txt": true,
    "robots-page": true,
    "robots-gptbot": false,
    "robots-claudebot": false,
    "robots-perplexitybot": true,
    "robots-google-extended": true,
    "llms-txt": true,
  });
  expect([cat.score, cat.passed, cat.total]).toEqual([null, 5, 7]);
});

test("siteCategory without robots or llms", () => {
  const status = Object.fromEntries(siteCategory("/", null, null).checks.map((c) => [c.id, c.pass]));
  expect([status["robots-txt"], status["llms-txt"], status["robots-page"], status["robots-gptbot"]]).toEqual([
    false,
    false,
    true,
    true,
  ]);
});

test("robotsAllows ignores patterns with more than 10 wildcards", () => {
  expect(robotsAllows(`User-agent: *\nDisallow: /${"*".repeat(10)}\n`, "Googlebot", "/x")).toBe(false);
  expect(robotsAllows(`User-agent: *\nDisallow: /${"*".repeat(11)}\n`, "Googlebot", "/x")).toBe(true);
});
