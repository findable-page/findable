export function fixedWindowLimiter(limit: number, windowMs: number, now: () => number = Date.now) {
  let windowStart = now();
  const counts = new Map<string, number>();
  return (key: string): boolean => {
    const t = now();
    if (t - windowStart >= windowMs) {
      windowStart = t;
      counts.clear();
    }
    const n = (counts.get(key) ?? 0) + 1;
    counts.set(key, n);
    return n <= limit;
  };
}
