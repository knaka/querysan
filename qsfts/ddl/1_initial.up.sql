CREATE VIRTUAL TABLE file_texts USING fts4(
    title TEXT,
    body TEXT,
    tokenize=unicode61
);

CREATE TABLE files(
    path text PRIMARY KEY NOT NULL,
    text_id int NOT NULL,
    updated_at datetime NOT NULL
);

CREATE INDEX index_files_text_id ON files(text_id);
