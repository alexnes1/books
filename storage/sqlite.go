package storage

import (
	"database/sql"
	_ "embed"
	"fmt"

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

func (s *SQLiteStorage) StoreBook(b Book) error {
	for _, a := range b.Authors {
		id, err := s.storeAuthor(a)
		fmt.Printf("Stored author [%s] with ID [%d] and ERROR [%s]\n", a, id, err)
	}
	return nil
}
