package receipt

type PrintTarget interface {
	Open() error
	Write([]byte) (int, error)
	Close() error
}
