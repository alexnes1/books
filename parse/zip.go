package parse

import (
	"archive/zip"
	"io"
	"log"
)

func walkZipArchive(root string, relativePath string, bookHandler *BookHandler, bookChan chan<- rawBook) {
	fullPath := getFullPath(root, relativePath)
	r, err := zip.OpenReader(fullPath)
	if err != nil {
		log.Println(err)
		return
	}

	defer r.Close()

	for _, f := range r.File {
		if bookHandler.canProcess(f.Name) {
			rc, err := f.Open()
			if err != nil {
				log.Printf("can't read %s", getFullPath(root, relativePath, f.Name))
				continue
			}

			bytes, err := io.ReadAll(rc)
			if err != nil {
				rc.Close()
				log.Printf("can't read %s", getFullPath(root, relativePath, f.Name))
				continue
			}
			book := rawBook{
				root:         root,
				relativePath: relativePath,
				filename:     f.Name,
				data:         bytes,
			}
			bookChan <- book

			rc.Close()
		}
	}
}
