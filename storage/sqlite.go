package storage

import (
	"database/sql"
	_ "embed"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var sqliteSchema string

func InitSqliteDb(dbFilePath string) (SQLiteStorage, error) {
	db, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		return SQLiteStorage{}, err
	}

	if _, err := db.Exec(sqliteSchema); err != nil {
		return SQLiteStorage{}, err
	}

	return SQLiteStorage{db: db}, nil
}

type SQLiteStorage struct {
	db *sql.DB
}

func (s *SQLiteStorage) storeAuthor(a Author) (int, error) {
	var id int
	err := s.db.QueryRow(`INSERT INTO authors(first_name, middle_name, last_name, nickname, homepage, email)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (first_name, middle_name, last_name) DO UPDATE SET
		nickname=excluded.nickname,
		homepage=excluded.homepage,
		email=excluded.email
		RETURNING id;`,
		a.FirstName, a.MiddleName, a.LastName, a.Nickname, a.Homepage, a.Email,
	).Scan(&id)
	return id, err
}

func (s *SQLiteStorage) storeGenre(g Genre) (int, error) {
	var id int
	err := s.db.QueryRow(`INSERT INTO genres(name) VALUES (?)
		ON CONFLICT (name) DO UPDATE SET name=excluded.name
		RETURNING id;`,
		g.Name,
	).Scan(&id)
	return id, err
}

func (s *SQLiteStorage) storeKeyword(k Keyword) (int, error) {
	var id int
	err := s.db.QueryRow(`INSERT INTO keywords(name) VALUES (?)
	ON CONFLICT (name) DO UPDATE SET name=excluded.name
		RETURNING id;`,
		k.Name,
	).Scan(&id)
	return id, err
}

func (s *SQLiteStorage) StoreBook(b Book) error {
	authorIds := []int{}
	for _, a := range b.Authors {
		id, err := s.storeAuthor(a)
		if err != nil {
			return err
		}
		authorIds = append(authorIds, id)
		// fmt.Printf("Stored author [%s] with ID [%d] and ERROR [%s]\n", a, id, err)
	}

	genreIds := []int{}
	for _, g := range b.Genres {
		id, err := s.storeGenre(g)
		if err != nil {
			return err
		}
		genreIds = append(genreIds, id)
		// fmt.Printf("Stored genre [%s] with ID [%d] and ERROR [%s]\n", g, id, err)
	}

	keywordIds := []int{}
	for _, k := range b.Keywords {
		id, err := s.storeKeyword(k)
		if err != nil {
			return err
		}
		keywordIds = append(keywordIds, id)
		// fmt.Printf("Stored keyword [%s] with ID [%d] and ERROR [%s]\n", k, id, err)
	}

	var bookId int
	err := s.db.QueryRow(`INSERT INTO books(
			title, annotation, bookdate, 
			publisher, publish_city, publish_year, ISBN, lang, src_lang, 
			root_path, relative_path, filename, filesize)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id;`,
		b.Title, b.Annotation, b.Date, b.Publisher, b.PublishCity, b.PublishYear,
		b.ISBN, b.Lang, b.SrcLang, b.File.RootPath, b.File.RelativePath, b.File.Filename, b.File.Filesize,
	).Scan(&bookId)
	if err != nil {
		return err
	}

	if len(authorIds) > 0 {
		for _, id := range authorIds {
			_, err := s.db.Exec(`INSERT INTO books_authors (book_id, author_id) VALUES (?, ?);`, bookId, id)
			if err != nil {
				return err
			}
		}
	}

	if len(genreIds) > 0 {
		for _, id := range genreIds {
			_, err := s.db.Exec(`INSERT INTO books_genres (book_id, genre_id) VALUES (?, ?);`, bookId, id)
			if err != nil {
				return err
			}
		}
	}

	if len(genreIds) > 0 {
		for _, id := range keywordIds {
			_, err := s.db.Exec(`INSERT INTO books_keywords (book_id, keyword_id) VALUES (?, ?);`, bookId, id)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
