CREATE VIRTUAL TABLE fileinfo USING fts3(
  path TEXT PRIMARY,
  title TEXT,
  words TEXT,
  updated_at TEXT,
);
