package main

import (
	"fmt"
	"os"

	"github.com/alexnes1/books/parse"
	"github.com/alexnes1/books/storage"
)

func main() {
	db, err := storage.InitSqliteDb("./booksdb.sqlite")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: can not initialize db (%s).\n", err)
		os.Exit(1)
	}

	parse.AddBooksInfoToDatabase(".", 50, 16, &db)
}
