import { guardedFetch, type SelfRoute } from "./guard";
import type { Check } from "./index";

const AI_CRAWLERS = ["GPTBot", "ClaudeBot", "PerplexityBot", "Google-Extended"];
const PAGE_CRAWLER = "*";
export const MAX_SITE_FILE_BYTES = 512 * 1024;
const MAX_ROBOTS_WILDCARDS = 10;

interface RobotsRule {
  allow: boolean;
  pattern: string;
}

interface RobotsGroup {
  agents: string[];
  rules: RobotsRule[];
}

export interface SiteCategory {
  score: null;
  passed: number;
  total: number;
  checks: Check[];
}

function parseRobots(robots: string): RobotsGroup[] {
  const groups: RobotsGroup[] = [];
  let inAgents = false;
  for (const raw of robots.split("\n")) {
    const line = raw.split("#")[0] ?? "";
    const colon = line.indexOf(":");
    if (colon < 0) continue;
    const key = line.slice(0, colon).trim().toLowerCase();
    const value = line.slice(colon + 1).trim();
    const current = groups[groups.length - 1];
    if (key === "user-agent") {
      const token = (value.split("/")[0] ?? "").trim().toLowerCase();
      if (!inAgents || !current) groups.push({ agents: [token], rules: [] });
      else current.agents.push(token);
      inAgents = true;
    } else if (key === "allow" || key === "disallow") {
      inAgents = false;
      if (current && value) current.rules.push({ allow: key === "allow", pattern: value });
    }
  }
  return groups;
}

function rulesFor(groups: RobotsGroup[], agent: string): RobotsRule[] {
  const name = agent.toLowerCase();
  const named = groups.filter((g) => g.agents.includes(name)).flatMap((g) => g.rules);
  if (groups.some((g) => g.agents.includes(name))) return named;
  return groups.filter((g) => g.agents.includes(PAGE_CRAWLER)).flatMap((g) => g.rules);
}

function patternMatches(pattern: string, path: string): boolean {
  if (pattern.split("*").length - 1 > MAX_ROBOTS_WILDCARDS) return false;
  const anchored = pattern.endsWith("$");
  const body = anchored ? pattern.slice(0, -1) : pattern;
  const expr = body
    .split("*")
    .map((part) => part.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"))
    .join(".*");
  return new RegExp(`^${expr}${anchored ? "$" : ""}`).test(path);
}

export function robotsAllows(robots: string, agent: string, path: string): boolean {
  let best: RobotsRule = { allow: true, pattern: "" };
  let bestLen = -1;
  for (const rule of rulesFor(parseRobots(robots), agent)) {
    if (!patternMatches(rule.pattern, path)) continue;
    const len = rule.pattern.length;
    if (len > bestLen || (len === bestLen && rule.allow)) {
      best = rule;
      bestLen = len;
    }
  }
  return best.allow;
}

export function siteCategory(pagePath: string, robots: string | null, llms: string | null): SiteCategory {
  const robotsText = robots ?? "";
  const checks: Check[] = [
    { id: "robots-txt", label: "robots.txt is present", pass: robots !== null },
    {
      id: "robots-page",
      label: "robots.txt lets crawlers fetch this page",
      pass: robotsAllows(robotsText, PAGE_CRAWLER, pagePath),
    },
    ...AI_CRAWLERS.map((bot) => ({
      id: `robots-${bot.toLowerCase()}`,
      label: `robots.txt lets ${bot} fetch this page`,
      pass: robotsAllows(robotsText, bot, pagePath),
    })),
    { id: "llms-txt", label: "/llms.txt is present", pass: llms !== null },
  ];
  return { score: null, passed: checks.filter((c) => c.pass).length, total: checks.length, checks };
}

async function optionalText(target: string, self?: SelfRoute): Promise<string | null> {
  try {
    const page = await guardedFetch(target, {}, self);
    if (page.status !== 200) return null;
    return page.body.slice(0, MAX_SITE_FILE_BYTES);
  } catch {
    return null;
  }
}

export async function siteChecks(pageUrl: string, self?: SelfRoute): Promise<SiteCategory> {
  const u = new URL(pageUrl);
  const [robots, llms] = await Promise.all([
    optionalText(`${u.origin}/robots.txt`, self),
    optionalText(`${u.origin}/llms.txt`, self),
  ]);
  return siteCategory(u.pathname + u.search, robots, llms);
}
