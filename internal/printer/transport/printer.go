package transport

type Printer interface {
	Open() error
	Write(data []byte) (int, error)
	Close() error
}
