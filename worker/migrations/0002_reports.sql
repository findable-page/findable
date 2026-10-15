CREATE TABLE IF NOT EXISTS reports (
  id TEXT PRIMARY KEY,
  host TEXT NOT NULL,
  url TEXT NOT NULL,
  created_at TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  json TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS reports_expires_at ON reports (expires_at);
