import { expect, test } from "bun:test";
import { readdirSync } from "node:fs";

const PUBLIC_DIR = new URL("../public/", import.meta.url);
const INDEX_HTML = new URL("../../internal/server/web/index.html", import.meta.url);

test("worker/public holds exactly the static assets the UI references", async () => {
  const html = await Bun.file(INDEX_HTML).text();
  const refs = [...html.matchAll(/<link rel="(?:icon|apple-touch-icon|manifest)" href="\/([^"]+)"/g)].map(
    (m) => m[1] ?? "",
  );
  const og = html.match(/property="og:image" content="https:\/\/[^/"]+\/([^"]+)"/)?.[1] ?? "";
  const demo = [...html.matchAll(/(?:src|poster)="\/(demo\.[a-z0-9]+)"/g)].map((m) => m[1] ?? "");
  const referenced = [...refs, og, ...demo].sort();
  expect(referenced).toEqual([
    "apple-touch-icon.png",
    "demo.jpg",
    "demo.mp4",
    "favicon.ico",
    "favicon.svg",
    "og.jpg",
    "site.webmanifest",
  ]);
  expect(readdirSync(PUBLIC_DIR.pathname).sort()).toEqual(
    [...referenced, "favicon-32.png", "icon-192.png", "icon-512.png", "og.png"].sort(),
  );
});
