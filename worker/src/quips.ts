// Bands by overall score: 90+, 70+, 50+, 30+, below 30.
type Lines = readonly [string, ...string[]];
export const QUIPS: readonly [Lines, Lines, Lines, Lines, Lines] = [
  [
    "Google and ChatGPT can't miss it.",
    "The robots have it on speed dial.",
    "Basically a celebrity to search engines.",
    "AI quotes it before you finish typing.",
    "Search engines send fan mail.",
    "Front page energy.",
    "ChatGPT's favourite bookmark.",
    "Findable from space.",
  ],
  [
    "Almost famous with the robots.",
    "Google likes it. ChatGPT is warming up.",
    "One good fix away from fame.",
    "Solid. The robots nod politely.",
    "Top of page two, emotionally.",
    "AI knows its name, not its number.",
    "Close enough to taste the snippet.",
    "B+ student, robot edition.",
  ],
  [
    "Half-visible to Google and AI.",
    "The robots squint and shrug.",
    "Google found it. ChatGPT is still looking.",
    "Visible, if you know where to look.",
    "Hiding in plain sight, mostly.",
    'The search bar\'s "maybe" pile.',
    "Halfway out of the shadows.",
    "AI has heard rumours about it.",
  ],
  [
    "Playing hide and seek with Google and AI.",
    "Winning hide and seek. That's the problem.",
    "Page 7 of Google sends its regards.",
    "ChatGPT has never heard of it.",
    "The robots walked right past.",
    "A well-kept secret from search engines.",
    "Lost somewhere behind the ads.",
    "Ghosted by the algorithm.",
  ],
  [
    "Invisible. Like, ninja-level invisible.",
    "Witness protection for websites.",
    "Search engines think it's a myth.",
    "Even its own sitemap can't find it.",
    "The robots filed a missing person report.",
    "Off the grid, by accident.",
    "Google asked: who?",
    "AI says it doesn't exist. Rude, but fair.",
  ],
];

function band(overall: number): 0 | 1 | 2 | 3 | 4 {
  if (overall >= 90) return 0;
  if (overall >= 70) return 1;
  if (overall >= 50) return 2;
  if (overall >= 30) return 3;
  return 4;
}

function hash(text: string): number {
  let h = 2166136261;
  for (let i = 0; i < text.length; i++) h = Math.imul(h ^ text.charCodeAt(i), 16777619);
  return h >>> 0;
}

export function scoreQuip(overall: number, host: string): string {
  const lines = QUIPS[band(overall)];
  return lines[hash(host) % lines.length] ?? lines[0];
}

const CHALLENGES: Lines = [
  "Can your site beat it? Free check, no signup.",
  "Bet yours scores lower. Prove me wrong, free.",
  "Your turn. 30 seconds, no signup.",
  "Check yours before your competitor does.",
  "Is your site hiding from ChatGPT too? Free check.",
  "Think you can top it? Free, no signup.",
  "See what Google and AI see on your site. Free.",
  "Run yours. Screenshot it. Brag or fix.",
];

export function shareDescription(s: { seo: number; aeo: number; geo: number }, host: string): string {
  const line = CHALLENGES[hash(`d:${host}`) % CHALLENGES.length] ?? CHALLENGES[0];
  return `SEO ${s.seo} · AEO ${s.aeo} · GEO ${s.geo}. ${line}`;
}

export interface CardTheme {
  background: string;
  accent: string;
}

const THEMES: readonly [CardTheme, ...CardTheme[]] = [
  { background: "linear-gradient(180deg,#fff4e6,#fffaf3 60%,#ffffff)", accent: "#4F2BFF" },
  { background: "linear-gradient(135deg,#eef2ff,#f8f5ff 60%,#ffffff)", accent: "#4F2BFF" },
  { background: "linear-gradient(135deg,#e8fff3,#f4fffa 60%,#ffffff)", accent: "#0E7C4A" },
  { background: "linear-gradient(135deg,#fff0eb,#fff8f5 60%,#ffffff)", accent: "#D1480D" },
  { background: "linear-gradient(135deg,#fff9db,#fffdf0 60%,#ffffff)", accent: "#7A4300" },
  { background: "linear-gradient(135deg,#ffeef8,#fff7fc 60%,#ffffff)", accent: "#B0127A" },
];

export function cardTheme(host: string): CardTheme {
  return THEMES[hash(`t:${host}`) % THEMES.length] ?? THEMES[0];
}
