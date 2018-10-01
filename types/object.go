package types

type Object interface {
	AsString() (*String, error)
	AsNumber() (*Number, error)
}

type InvalidCastError struct{}

func (e InvalidCastError) Error() string {
	return "invalid cast"
}
