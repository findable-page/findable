export const MAX_PAGE_BYTES = 5 * 1024 * 1024;
export const FETCH_TIMEOUT_MS = 10_000;
const MAX_REDIRECTS = 5;
const BLOCKED_SUFFIXES = [".localhost", ".local", ".internal", ".home.arpa"];
const BLOCKED_V4: [number, number][] = [
  [0x00000000, 8],
  [0x0a000000, 8],
  [0x64400000, 10],
  [0x7f000000, 8],
  [0xa9fe0000, 16],
  [0xac100000, 12],
  [0xc0000000, 24],
  [0xc0a80000, 16],
  [0xc6120000, 15],
  [0xc0000200, 24],
  [0xc6336400, 24],
  [0xcb007100, 24],
  [0xe0000000, 3],
];

export class BlockedTargetError extends Error {
  constructor(target: string) {
    super(`blocked address: only public internet hosts can be scanned (${target})`);
    this.name = "BlockedTargetError";
  }
}

export interface FetchedPage {
  url: string;
  status: number;
  headers: Headers;
  body: string;
}

function ipv4Number(host: string): number | null {
  const parts = host.split(".");
  if (parts.length !== 4 || !parts.every((p) => /^\d{1,3}$/.test(p) && Number(p) <= 255)) return null;
  return parts.reduce((n, p) => n * 256 + Number(p), 0);
}

function publicV4(n: number): boolean {
  return !BLOCKED_V4.some(([base, bits]) => Math.floor(n / 2 ** (32 - bits)) === Math.floor(base / 2 ** (32 - bits)));
}

function publicV6(host: string): boolean {
  const mapped = /^::ffff:([0-9a-f]{1,4}):([0-9a-f]{1,4})$/.exec(host);
  if (mapped) return publicV4(Number.parseInt(mapped[1] ?? "0", 16) * 65536 + Number.parseInt(mapped[2] ?? "0", 16));
  const head = (host.split("::")[0] ?? "").split(":");
  const [first = 0, second = 0] = [0, 1].map((i) => Number.parseInt(head[i] || "0", 16));
  if (first === 0x2002 || (first === 0x2001 && (second === 0 || second === 0xdb8))) return false;
  if (first === 0x3fff && second < 0x1000) return false;
  return first >= 0x2000 && first <= 0x3fff;
}

function parseURL(target: string): URL | null {
  try {
    return new URL(target);
  } catch {
    return null;
  }
}

export function publicTarget(target: string): boolean {
  const u = parseURL(target);
  if (!u) return false;
  if (u.protocol !== "http:" && u.protocol !== "https:") return false;
  if (u.port !== "" && u.port !== "80" && u.port !== "443") return false;
  const host = u.hostname.toLowerCase().replace(/\.+$/, "");
  if (host.startsWith("[")) return publicV6(host.slice(1, -1));
  const v4 = ipv4Number(host);
  if (v4 !== null) return publicV4(v4);
  if (host === "localhost" || !host.includes(".")) return false;
  return !BLOCKED_SUFFIXES.some((s) => host.endsWith(s));
}

async function readCapped(resp: Response): Promise<string> {
  if (!resp.body) return "";
  const reader = resp.body.getReader();
  const chunks: Uint8Array[] = [];
  let size = 0;
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    size += value.byteLength;
    if (size > MAX_PAGE_BYTES) {
      await reader.cancel();
      throw new Error(`page is larger than ${MAX_PAGE_BYTES} bytes`);
    }
    chunks.push(value);
  }
  const all = new Uint8Array(size);
  let offset = 0;
  for (const c of chunks) {
    all.set(c, offset);
    offset += c.byteLength;
  }
  return new TextDecoder().decode(all);
}

export interface SelfRoute {
  hosts: string[];
  handle: (request: Request) => Promise<Response>;
}

async function fetchHop(current: string, headers: Record<string, string>, signal: AbortSignal, self?: SelfRoute) {
  const u = new URL(current);
  if (!self?.hosts.includes(u.hostname.replace(/\.+$/, "").toLowerCase()))
    return fetch(current, { headers, redirect: "manual", signal });
  if (u.pathname.startsWith("/api/")) throw new BlockedTargetError(current);
  return self.handle(new Request(current, { method: "GET", headers }));
}

export async function guardedFetch(
  target: string,
  headers: Record<string, string> = {},
  self?: SelfRoute,
): Promise<FetchedPage> {
  const signal = AbortSignal.timeout(FETCH_TIMEOUT_MS);
  let current = target;
  for (let hop = 0; hop <= MAX_REDIRECTS; hop++) {
    if (!publicTarget(current)) throw new BlockedTargetError(current);
    const resp = await fetchHop(current, headers, signal, self);
    const location = resp.headers.get("location");
    if (resp.status >= 300 && resp.status < 400 && location) {
      await resp.body?.cancel();
      current = new URL(location, current).toString();
      continue;
    }
    return { url: current, status: resp.status, headers: resp.headers, body: await readCapped(resp) };
  }
  throw new Error("too many redirects");
}
