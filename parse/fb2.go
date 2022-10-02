package parse

import (
	"strings"

	"github.com/alexnes1/books/storage"
	"github.com/centrypoint/fb2"
)

func processFB2(raw rawBook) (storage.Book, error) {
	parser := fb2.New(raw.data)
	metadata, err := parser.Unmarshal()
	if err != nil {
		return storage.Book{}, err
	}

	b := storage.Book{}
	b.Title = metadata.Description.TitleInfo.BookTitle
	b.Annotation = metadata.Description.TitleInfo.Annotation
	b.Date = metadata.Description.TitleInfo.Date
	b.Publisher = metadata.Description.PublishInfo.Publisher
	b.PublishCity = metadata.Description.PublishInfo.City
	b.PublishYear = metadata.Description.PublishInfo.Year
	b.ISBN = metadata.Description.PublishInfo.ISBN
	b.Lang = metadata.Description.TitleInfo.Lang
	b.SrcLang = metadata.Description.TitleInfo.SrcLang
	b.File = storage.FileInfo{
		RootPath:     raw.root,
		RelativePath: raw.relativePath,
		Filename:     raw.filename,
		Filesize:     raw.filesize,
	}

	for _, a := range metadata.Description.TitleInfo.Author {
		b.Authors = append(b.Authors, storage.Author{
			FirstName:  a.FirstName,
			LastName:   a.LastName,
			MiddleName: a.MiddleName,
			Nickname:   a.Nickname,
			Homepage:   a.HomePage,
			Email:      a.Email,
		})
	}

	for _, g := range metadata.Description.TitleInfo.Genre {
		b.Genres = append(b.Genres, storage.Genre{Name: g})
	}

	for _, k := range strings.Split(metadata.Description.TitleInfo.Keywords, ",") {
		b.Keywords = append(b.Keywords, storage.Keyword{Name: strings.Trim(k, " ")})
	}

	return b, nil
}
