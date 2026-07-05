package domain

type Storage interface {
	SaveObject(name string, data []byte) error
	GetObject(name string) ([]byte, error)
	DeleteObject(name string) error
}
