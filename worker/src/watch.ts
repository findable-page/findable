const EMAIL = /^[^\s@]{1,64}@[^\s@]+\.[a-z]{2,}$/;
const HOST = /^[a-z0-9.-]+\.[a-z]{2,}$/;

export interface WatchSignup {
  email: string;
  host: string;
}

export type WatchParse = WatchSignup | "spam" | string;

export function parseWatch(body: unknown): WatchParse {
  const b = (body ?? {}) as Record<string, unknown>;
  if (typeof b.hp === "string" && b.hp !== "") return "spam";
  const email = typeof b.email === "string" ? b.email.trim().toLowerCase() : "";
  if (email.length > 254 || !EMAIL.test(email)) return "enter a valid email";
  if (b.consent !== true) return "tick the box to agree to score emails";
  const host = hostOf(typeof b.url === "string" ? b.url : "");
  if (!host) return "scan a site first";
  return { email, host };
}

function hostOf(raw: string): string | null {
  try {
    const host = new URL(/^https?:\/\//i.test(raw) ? raw : `https://${raw}`).hostname.toLowerCase();
    return HOST.test(host) ? host : null;
  } catch {
    return null;
  }
}

export async function saveWatch(db: D1Database, s: WatchSignup): Promise<void> {
  await db
    .prepare("INSERT OR IGNORE INTO watchers (email, host, created_at) VALUES (?, ?, ?)")
    .bind(s.email, s.host, new Date().toISOString())
    .run();
}
