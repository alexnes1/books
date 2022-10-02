package main

import (
	"github.com/alexnes1/books/parse"
)

func main() {
	// addBooksInfoToDatabase(".", 20, 4)
	// storage.PrintSchema()
	parse.AddBooksInfoToDatabase(".", 20, 4)
}
