import { expect, test } from "bun:test";
import { readdirSync } from "node:fs";

const PUBLIC_DIR = new URL("../public/", import.meta.url);
const INDEX_HTML = new URL("../../internal/server/web/index.html", import.meta.url);

test("worker/public holds exactly the static assets the UI references", async () => {
  const html = await Bun.file(INDEX_HTML).text();
  const refs = [...html.matchAll(/<link rel="(?:icon|apple-touch-icon)" href="\/([^"]+)"/g)].map((m) => m[1] ?? "");
  const og = html.match(/property="og:image" content="https:\/\/[^/"]+\/([^"]+)"/)?.[1] ?? "";
  const referenced = [...refs, og].sort();
  expect(referenced).toEqual(["apple-touch-icon.png", "favicon.svg", "og.png"]);
  expect(readdirSync(PUBLIC_DIR).sort()).toEqual(referenced);
});
