CREATE TABLE IF NOT EXISTS watchers (
  email TEXT NOT NULL,
  host TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (email, host)
);
