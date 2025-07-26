package auth

type ErrorKind uint8

var (
	ErrKeyNotFound  = &KeyError{Kind: KeyNotFound, Message: "key not found"}
	ErrKeyDisabled  = &KeyError{Kind: KeyDisabled, Message: "key is disabled"}
	ErrKeyExpired   = &KeyError{Kind: KeyExpired, Message: "key has expired"}
	ErrKeyMalformed = &KeyError{Kind: KeyMalformed, Message: "malformed key"}
)

const (
	KeyNotFound ErrorKind = iota
	KeyDisabled
	KeyExpired
	KeyMalformed
	InternalError
	KeyCorrupted
)

type KeyError struct {
	Kind    ErrorKind
	Message string
	cause   error
}

func (e *KeyError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	switch e.Kind {
	case KeyNotFound:
		return "key not found"
	case KeyDisabled:
		return "key disabled"
	case KeyExpired:
		return "key expired"
	case KeyMalformed:
		return "key malformed"
	case InternalError:
		return "internal error"
	case KeyCorrupted:
		return "api key corrupted"
	default:
		return "unknown error"
	}
}

func (e *KeyError) Is(target error) bool {
	t, ok := target.(*KeyError)
	return ok && e.Kind == t.Kind
}

func (e *KeyError) Unwrap() error {
	return e.cause
}

func NewKeyError(kind ErrorKind, msg string) *KeyError {
	return &KeyError{Kind: kind, Message: msg}
}

func NewKeyErrorWithCause(kind ErrorKind, msg string, cause error) *KeyError {
	return &KeyError{Kind: kind, Message: msg, cause: cause}
}
