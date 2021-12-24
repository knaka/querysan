CREATE VIRTUAL TABLE file USING fts3(
  path TEXT,
  title TEXT,
  words TEXT,
  timestamp DATE
);
