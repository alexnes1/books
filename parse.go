package main

import (
	"archive/zip"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/centrypoint/fb2"
)

type extensionHandlerFunc func(root string, relativePath string, bookHandler *BookHandler, books chan<- rawBook)

const queueSize = 50
const numWorkers = 4

type NonBookHandler struct {
	extensions map[string]extensionHandlerFunc
}

func NewNonBookHandler() *NonBookHandler {
	return &NonBookHandler{extensions: map[string]extensionHandlerFunc{}}
}

func (h *NonBookHandler) register(extension string, f extensionHandlerFunc) {
	h.extensions[strings.ToUpper(extension)] = f
}

func (h *NonBookHandler) process(root string, relativePath string, bookHandler *BookHandler, books chan<- rawBook) {
	fileExt := strings.ToUpper(filepath.Ext(relativePath))
	f, ok := h.extensions[fileExt]
	if !ok {
		return
	}
	f(root, relativePath, bookHandler, books)
}

type rawBook struct {
	root         string
	relativePath string
	filename     string
	data         []byte
}

func (b *rawBook) FullPath() string {
	return getFullPath(b.root, b.relativePath, b.filename)
}

type bookHandlerFunc func(rawBook)

type BookHandler struct {
	extensions map[string]bookHandlerFunc
}

func NewBookHandler() *BookHandler {
	return &BookHandler{extensions: map[string]bookHandlerFunc{}}
}

func (h *BookHandler) register(extension string, f bookHandlerFunc) {
	h.extensions[strings.ToUpper(extension)] = f
}

func (h *BookHandler) canProcess(path string) bool {
	fileExt := strings.ToUpper(filepath.Ext(path))
	_, ok := h.extensions[fileExt]
	return ok
}

func (h *BookHandler) process(data rawBook) {
	fileExt := strings.ToUpper(filepath.Ext(data.filename))
	f, ok := h.extensions[fileExt]
	if !ok {
		return
	}
	f(data)
}

func processFB2(raw rawBook) {
	parser := fb2.New(raw.data)
	metadata, err := parser.Unmarshal()
	if err != nil {
		log.Printf("%s: %s", raw.FullPath(), err)
		return
	}
	log.Printf("%s: OK [%s -- %v]\n", raw.FullPath(), metadata.Description.TitleInfo.BookTitle,
		metadata.Description.PublishInfo.BookName)
	// metadata.Description.TitleInfo.SrcLang)
}

func getFullPath(parts ...string) string {
	return strings.Join(parts, string(os.PathSeparator))
}

func getFileName(path string) string {
	parts := strings.Split(path, string(os.PathSeparator))
	return parts[len(parts)-1]
}

func getPathWithoutFilename(path string) string {
	parts := strings.Split(path, string(os.PathSeparator))
	if len(parts) == 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], string(os.PathSeparator))
}

func listBookfiles(root string, bookHandler *BookHandler, nonBookHandler *NonBookHandler, bookChan chan<- rawBook, wg *sync.WaitGroup) {
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if bookHandler.canProcess(path) {
			bytes, err := os.ReadFile(path)
			if err != nil {
				log.Printf("can't read %s", getFullPath(root, path))
			}
			book := rawBook{
				root:         root,
				relativePath: getPathWithoutFilename(path),
				filename:     getFileName(path),
				data:         bytes,
			}
			bookChan <- book
		} else {
			nonBookHandler.process(root, path, bookHandler, bookChan)
		}

		return nil
	})

	if err != nil {
		log.Println(err)
	}

	close(bookChan)
	wg.Done()
}

func processRawBook(bookChan <-chan rawBook, bookHandler *BookHandler, wg *sync.WaitGroup) {
	for book := range bookChan {
		bookHandler.process(book)
	}
	wg.Done()
}

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

func addBooksInfoToDatabase(dir string) {
	nbh := NewNonBookHandler()
	nbh.register(".zip", walkZipArchive)

	bh := NewBookHandler()
	bh.register(".fb2", processFB2)

	books := make(chan rawBook, queueSize)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go listBookfiles(dir, bh, nbh, books, wg)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go processRawBook(books, bh, wg)
	}
	wg.Wait()
}
