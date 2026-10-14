import { expect, test } from "bun:test";
import worker from "../src/index";

const ORIGIN = "https://findable.piply.xyz";
const FLOORS = { seo: 90, aeo: 85, geo: 80 };

interface Category {
  score: number;
  checks: { id: string; pass: boolean }[];
}

test("the home page scores SEO 90, AEO 85 and GEO 80 or better when scanned in-process", async () => {
  const resp = await worker.fetch(
    new Request(`${ORIGIN}/api/scan`, {
      method: "POST",
      headers: { "CF-Connecting-IP": "203.0.113.200" },
      body: JSON.stringify({ url: `${ORIGIN}/` }),
    }),
  );
  const scan = (await resp.json()) as Record<keyof typeof FLOORS, Category>;
  for (const [name, floor] of Object.entries(FLOORS) as [keyof typeof FLOORS, number][]) {
    const failing = scan[name].checks.filter((c) => !c.pass).map((c) => c.id);
    expect({ name, score: scan[name].score, failing }).toEqual({ name, score: expect.any(Number), failing });
    expect(scan[name].score).toBeGreaterThanOrEqual(floor);
  }
  const passed = new Set([...scan.seo.checks, ...scan.aeo.checks, ...scan.geo.checks].filter((c) => c.pass).map((c) => c.id));
  for (const id of ["jsonld", "aeo_structured_data", "geo_author_schema", "geo_org_schema"]) expect(passed.has(id)).toBe(true);
});
