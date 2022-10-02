package parse

import (
	"strings"

	"github.com/alexnes1/books/fb2"
	"github.com/alexnes1/books/storage"
)

func processFB2(raw rawBook) (storage.Book, error) {
	parser := fb2.New(raw.data)
	metadata, err := parser.Unmarshal()
	if err != nil {
		return storage.Book{}, err
	}

	b := storage.Book{}
	b.Title = strings.Trim(metadata.Description.TitleInfo.BookTitle, " ")
	b.Annotation = strings.Trim(metadata.Description.TitleInfo.Annotation, " ")
	b.Date = strings.Trim(metadata.Description.TitleInfo.Date, " ")
	b.Publisher = strings.Trim(metadata.Description.PublishInfo.Publisher, " ")
	b.PublishCity = strings.Trim(metadata.Description.PublishInfo.City, " ")
	b.PublishYear = metadata.Description.PublishInfo.Year
	b.ISBN = strings.Trim(metadata.Description.PublishInfo.ISBN, " ")
	b.Lang = strings.Trim(metadata.Description.TitleInfo.Lang, " ")
	b.SrcLang = strings.Trim(metadata.Description.TitleInfo.SrcLang, " ")
	b.File = storage.FileInfo{
		RootPath:     raw.root,
		RelativePath: raw.relativePath,
		Filename:     raw.filename,
		Filesize:     raw.filesize,
	}

	for _, a := range metadata.Description.TitleInfo.Author {
		b.Authors = append(b.Authors, storage.Author{
			FirstName:  strings.Trim(a.FirstName, " "),
			LastName:   strings.Trim(a.LastName, " "),
			MiddleName: strings.Trim(a.MiddleName, " "),
			Nickname:   strings.Trim(a.Nickname, " "),
			Homepage:   strings.Trim(a.HomePage, " "),
			Email:      strings.Trim(a.Email, " "),
		})
	}

	for _, g := range metadata.Description.TitleInfo.Genre {
		b.Genres = append(b.Genres, storage.Genre{Name: strings.Trim(g, " ")})
	}

	for _, k := range strings.Split(metadata.Description.TitleInfo.Keywords, ",") {
		b.Keywords = append(b.Keywords, storage.Keyword{Name: strings.Trim(k, " ")})
	}

	return b, nil
}
