import { afterEach, expect, test } from "bun:test";
import { BlockedTargetError, guardedFetch, MAX_PAGE_BYTES, publicTarget } from "../src/guard";

const realFetch = globalThis.fetch;
afterEach(() => {
  globalThis.fetch = realFetch;
});

function mockFetch(handler: (url: string) => Response) {
  globalThis.fetch = (async (input: RequestInfo | URL) => handler(String(input))) as typeof fetch;
}

test("publicTarget accepts public http(s) hosts only", () => {
  const cases: [string, boolean][] = [
    ["https://example.com/a", true],
    ["http://8.8.8.8/", true],
    ["file:///etc/passwd", false],
    ["ftp://example.com/", false],
    ["http://localhost:8080/", false],
    ["http://api.localhost/", false],
    ["http://127.0.0.1/", false],
    ["http://10.0.0.5/", false],
    ["http://172.20.1.1/", false],
    ["http://192.168.68.60/", false],
    ["http://169.254.169.254/latest/meta-data", false],
    ["http://100.104.51.75/", false],
    ["http://0.0.0.0/", false],
    ["http://[::1]/", false],
    ["http://[fd00::1]/", false],
    ["http://[fe80::1]/", false],
    ["http://[::ffff:127.0.0.1]/", false],
    ["http://[2002:7f00:1::1]/", false],
    ["http://[2001::1]/", false],
    ["http://[2001:0:4136:e378::1]/", false],
    ["http://[2001:4860:4860::8888]/", true],
    ["http://metadata.google.internal/", false],
    ["http://printer.local/", false],
    ["http://localhost./", false],
    ["http://foo.internal./", false],
    ["http://printer.local../", false],
    ["https://example.com./", true],
    ["not a url", false],
  ];
  for (const [url, want] of cases) {
    expect([url, publicTarget(url)]).toEqual([url, want]);
  }
});

test("guardedFetch re-checks every redirect hop", async () => {
  mockFetch((url) =>
    url.startsWith("https://example.com")
      ? new Response(null, { status: 302, headers: { location: "http://169.254.169.254/" } })
      : new Response("secret"),
  );
  await expect(guardedFetch("https://example.com/")).rejects.toBeInstanceOf(BlockedTargetError);
});

test("guardedFetch follows public redirects and returns the final body", async () => {
  mockFetch((url) =>
    url === "https://example.com/"
      ? new Response(null, { status: 301, headers: { location: "/home" } })
      : new Response(`page at ${url}`),
  );
  const res = await guardedFetch("https://example.com/");
  expect([res.status, res.body]).toEqual([200, "page at https://example.com/home"]);
});

test("guardedFetch refuses bodies over the cap", async () => {
  mockFetch(() => new Response("a".repeat(MAX_PAGE_BYTES + 1)));
  await expect(guardedFetch("https://example.com/")).rejects.toThrow("larger than");
});
