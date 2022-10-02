package parse

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alexnes1/books/storage"
)

type rawBook struct {
	root         string
	relativePath string
	filename     string
	filesize     int
	data         []byte
}

func (b *rawBook) FullPath() string {
	return getFullPath(b.root, b.relativePath, b.filename)
}

type extensionHandlerFunc func(root string, relativePath string, bookHandler *BookHandler, books chan<- rawBook)

type bookHandlerFunc func(rawBook) (storage.Book, error)

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

func (h *BookHandler) process(data rawBook) (storage.Book, error) {
	fileExt := strings.ToUpper(filepath.Ext(data.filename))
	f, ok := h.extensions[fileExt]
	if !ok {
		return storage.Book{}, fmt.Errorf("unknown book format")
	}
	return f(data)
}

func walkBookfiles(root string, bookHandler *BookHandler, nonBookHandler *NonBookHandler, bookChan chan<- rawBook, wg *sync.WaitGroup) {
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if bookHandler.canProcess(path) {
			bytes, err := os.ReadFile(path)
			if err != nil {
				log.Printf("can't read %s", getFullPath(root, path))
				return nil
			}
			book := rawBook{
				root:         root,
				relativePath: getPathWithoutFilename(path),
				filename:     getFileName(path),
				filesize:     len(bytes),
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

func processRawBook(bookChan <-chan rawBook, bookHandler *BookHandler, storeChan chan<- storage.Book, wg *sync.WaitGroup) {
	for book := range bookChan {
		book, err := bookHandler.process(book)
		if err != nil {
			log.Println(err)
			continue
		}
		storeChan <- book
	}
	wg.Done()
}

func saveBooks(storeChan <-chan storage.Book, db storage.Storage, wg *sync.WaitGroup) {
	for book := range storeChan {
		log.Println(book.String())
		err := db.StoreBook(book)
		if err != nil {
			log.Println(err)
		}
	}
	wg.Done()
}

func AddBooksInfoToDatabase(dir string, queueSize int, numWorkers int, db storage.Storage) {
	nbh := NewNonBookHandler()
	nbh.register(".zip", walkZipArchive)
	nbh.register(".fbz", walkZipArchive)

	bh := NewBookHandler()
	bh.register(".fb2", processFB2)

	booksChan := make(chan rawBook, queueSize)
	storeChan := make(chan storage.Book)

	processGroup := &sync.WaitGroup{}
	processGroup.Add(1)
	go walkBookfiles(dir, bh, nbh, booksChan, processGroup)
	for i := 0; i < numWorkers; i++ {
		processGroup.Add(1)
		go processRawBook(booksChan, bh, storeChan, processGroup)
	}

	storeGroup := &sync.WaitGroup{}
	storeGroup.Add(1)
	go saveBooks(storeChan, db, storeGroup)

	processGroup.Wait()
	close(storeChan)
	storeGroup.Wait()

}
