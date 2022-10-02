package storage

import (
	"fmt"
	"os"
	"strings"
)

type Storage interface {
	StoreBook(b Book) error
}

type stringer interface {
	String() string
}

func sliceToString[T stringer](s []T) string {
	lst := make([]string, 0, len(s))

	for _, el := range s {
		lst = append(lst, el.String())
	}
	return strings.Join(lst, ", ")

}

type Author struct {
	id         int
	FirstName  string
	MiddleName string
	LastName   string
	Nickname   string
	Homepage   string
	Email      string
}

func (a Author) String() string {
	return fmt.Sprintf("%s %s", a.FirstName, a.LastName)
}

type Genre struct {
	id   int
	Name string
}

func (g Genre) String() string {
	return g.Name
}

type Keyword struct {
	id   int
	Name string
}

func (k Keyword) String() string {
	return k.Name
}

type FileInfo struct {
	RootPath     string
	RelativePath string
	Filename     string
	Filesize     int
}

func (fi FileInfo) String() string {
	return strings.Join([]string{fi.RootPath, fi.RelativePath, fi.Filename}, string(os.PathSeparator))
}

type Book struct {
	id          int
	Title       string
	Annotation  string
	Date        string
	Publisher   string
	PublishCity string
	PublishYear int
	ISBN        string
	Lang        string
	SrcLang     string
	Authors     []Author
	Genres      []Genre
	Keywords    []Keyword
	File        FileInfo
}

func (b *Book) AuthorsString() string {
	lst := make([]string, 0, len(b.Authors))
	for _, a := range b.Authors {
		lst = append(lst, a.String())
	}
	return strings.Join(lst, ", ")
}

func (b *Book) String() string {
	return fmt.Sprintf("Title: %s\nAuthors: %s\nGenres: %s\nKeywords: %s\nFile: %s\n\n\n",
		b.Title, sliceToString(b.Authors), sliceToString(b.Genres), sliceToString(b.Keywords), b.File)
}
