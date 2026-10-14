import type { Check } from "./index";

export interface CwvCategory {
  score: number;
  checks: Check[];
  passed: number;
  total: number;
}

interface Lighthouse {
  lighthouseResult?: {
    categories?: { performance?: { score?: number } };
    audits?: Record<string, { numericValue?: number }>;
  };
  error?: { message?: string };
}

const seconds = (ms: number) => `${(ms / 1000).toFixed(1)} s`;

const VITALS = [
  { id: "cwv_lcp", audit: "largest-contentful-paint", good: 2500, name: "Largest Contentful Paint", show: seconds },
  {
    id: "cwv_cls",
    audit: "cumulative-layout-shift",
    good: 0.1,
    name: "Cumulative Layout Shift",
    show: (v: number) => v.toFixed(2),
  },
  {
    id: "cwv_tbt",
    audit: "total-blocking-time",
    good: 200,
    name: "Total Blocking Time",
    show: (v: number) => `${Math.round(v)} ms`,
  },
  { id: "cwv_fcp", audit: "first-contentful-paint", good: 1800, name: "First Contentful Paint", show: seconds },
];

export function cwvFromLighthouse(body: unknown): CwvCategory {
  const lh = (body as Lighthouse).lighthouseResult;
  const perf = lh?.categories?.performance?.score;
  if (!lh?.audits || typeof perf !== "number") {
    throw new Error((body as Lighthouse).error?.message ?? "PageSpeed Insights returned no data");
  }
  const checks: Check[] = VITALS.map((v) => {
    const value = lh.audits?.[v.audit]?.numericValue ?? Number.POSITIVE_INFINITY;
    return {
      id: v.id,
      label: `${v.name} ${v.show(value)} (good: ${v.show(v.good)} or less)`,
      pass: value <= v.good,
      category: "cwv",
    };
  });
  const passed = checks.filter((c) => c.pass).length;
  return { score: Math.round(perf * 100), checks, passed, total: checks.length };
}

export async function pageSpeed(target: string, key: string): Promise<CwvCategory> {
  const api = new URL("https://www.googleapis.com/pagespeedonline/v5/runPagespeed");
  api.search = new URLSearchParams({ url: target, strategy: "mobile", category: "performance", key }).toString();
  const resp = await fetch(api);
  return cwvFromLighthouse(await resp.json().catch(() => ({})));
}
