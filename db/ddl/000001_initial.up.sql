CREATE VIRTUAL TABLE fileinfo USING fts3(
  path TEXT,
  title TEXT,
  words TEXT,
);
