CREATE TABLE IF NOT EXISTS books (
    id INTEGER PRIMARY KEY,
    title VARCHAR NOT NULL,
    annotation VARCHAR,
    bookdate VARCHAR,
    publisher VARCHAR,
    publish_city VARCHAR,
    publish_year INTEGER,
    ISBN VARCHAR,
    lang VARCHAR,
    src_lang VARCHAR,
    root_path VARCHAR,
    relative_path VARCHAR,
    filename VARCHAR,
    filesize INT,
    UNIQUE (root_path, relative_path, filename)
);
CREATE INDEX IF NOT EXISTS idx_books_title ON books (title);
CREATE INDEX IF NOT EXISTS idx_books_lang ON books (lang);
CREATE INDEX IF NOT EXISTS idx_books_publisher ON books (publisher);
CREATE INDEX IF NOT EXISTS idx_books_isbn ON books (ISBN);

CREATE TABLE IF NOT EXISTS authors (
    id INTEGER PRIMARY KEY,
    first_name VARCHAR,
	middle_name VARCHAR,
	last_name VARCHAR,
	nickname VARCHAR,
	homepage VARCHAR,
	email VARCHAR,
    UNIQUE (first_name, middle_name, last_name)
);
CREATE INDEX IF NOT EXISTS idx_authors_lname ON authors (last_name);
CREATE INDEX IF NOT EXISTS idx_authors_flname ON authors (first_name, last_name);

CREATE TABLE IF NOT EXISTS books_authors (
    id INTEGER PRIMARY KEY,
    book_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL,

    FOREIGN KEY (book_id) REFERENCES books (id),
    FOREIGN KEY (author_id) REFERENCES authors (id)
);

CREATE TABLE IF NOT EXISTS genres (
    id INTEGER PRIMARY KEY,
    name VARCHAR NOT NULL UNIQUE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_genre_name ON genres (name);

CREATE TABLE IF NOT EXISTS books_genres (
    id INTEGER PRIMARY KEY,
    book_id INTEGER NOT NULL,
    genre_id INTEGER NOT NULL,

    FOREIGN KEY (book_id) REFERENCES books (id),
    FOREIGN KEY (genre_id) REFERENCES genres (id)
);

CREATE TABLE IF NOT EXISTS keywords (
    id INTEGER PRIMARY KEY,
    name VARCHAR NOT NULL UNIQUE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_keyword_name ON keywords (name);

CREATE TABLE IF NOT EXISTS books_keywords (
    id INTEGER PRIMARY KEY,
    book_id INTEGER NOT NULL,
    keyword_id INTEGER NOT NULL,

    FOREIGN KEY (book_id) REFERENCES books (id),
    FOREIGN KEY (keyword_id) REFERENCES keywords (id)
);

