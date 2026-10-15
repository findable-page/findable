const SCANS = "scans";

export async function countScan(db: D1Database): Promise<void> {
  await db
    .prepare("INSERT INTO stats (name, value) VALUES (?, 1) ON CONFLICT(name) DO UPDATE SET value = value + 1")
    .bind(SCANS)
    .run();
}

export async function scanCount(db: D1Database): Promise<number> {
  const row = await db.prepare("SELECT value FROM stats WHERE name = ?").bind(SCANS).first<{ value: number }>();
  return row?.value ?? 0;
}
