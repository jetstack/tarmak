package kv

type NotFoundError struct {
	msg string // description of error
}

func (e *NotFoundError) Error() string { return e.msg }

func NewNotFoundError(msg string) *NotFoundError {
	return &NotFoundError{
		msg: msg,
	}
}

// Service defines a basic key-value store. Implementations of this interface
// may or may not guarantee consistency or security properties.
type Service interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Test(key string) error
}
