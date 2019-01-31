package types

type Object interface {
	AsString() (*String, error)
	AsNumber() (*Number, error)
	AsFunction() (*Function, error)
}

type InvalidCastError struct{}

func (e InvalidCastError) Error() string {
	return "invalid cast"
}
