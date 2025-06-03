package shortener

type URLStore interface {
	Save(shortURL, originalURL string) error
	Get(shortURL string) (string, error)
}

type URLStoreWithDBforHandler interface {
	URLStore
	GetByOriginalURL(originalURL string) (string, error)
}
