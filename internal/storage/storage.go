package storage

import "errors"

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url already exists")
)

// URLStorage defines the interface for URL storage operations
type URLStorage interface {
	SaveURL(urlToSave, alias string) (int64, error)
	GetURL(alias string) (string, error)
	DeleteURL(alias string) error
}
