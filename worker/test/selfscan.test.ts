import { expect, test } from "bun:test";
import worker from "../src/index";
import { learnPaths } from "../src/learn";

const ORIGIN = "https://findable.page";
const FLOORS = { seo: 90, aeo: 85, geo: 80 };

interface Category {
  score: number;
  checks: { id: string; pass: boolean }[];
}

const scanPaths = ["/", ...learnPaths()];

test.each(scanPaths)("%s scores SEO 90, AEO 85 and GEO 80 or better when scanned in-process", async (path) => {
  const resp = await worker.fetch(
    new Request(`${ORIGIN}/api/scan`, {
      method: "POST",
      headers: { "CF-Connecting-IP": `203.0.113.${100 + scanPaths.indexOf(path)}` },
      body: JSON.stringify({ url: `${ORIGIN}${path}` }),
    }),
  );
  const scan = (await resp.json()) as Record<keyof typeof FLOORS, Category>;
  for (const [name, floor] of Object.entries(FLOORS) as [keyof typeof FLOORS, number][]) {
    const failing = scan[name].checks.filter((c) => !c.pass).map((c) => c.id);
    expect({ path, name, score: scan[name].score, failing }).toEqual({
      path,
      name,
      score: expect.any(Number),
      failing,
    });
    expect(scan[name].score).toBeGreaterThanOrEqual(floor);
  }
  if (path !== "/") return;
  const passed = new Set(
    [...scan.seo.checks, ...scan.aeo.checks, ...scan.geo.checks].filter((c) => c.pass).map((c) => c.id),
  );
  for (const id of ["jsonld", "aeo_structured_data", "geo_author_schema", "geo_org_schema"])
    expect(passed.has(id)).toBe(true);
});
