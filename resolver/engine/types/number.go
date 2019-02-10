package types

type Number float64

func NewNumber(v float64) *Number {
	n := Number(v)
	return &n
}

func (n *Number) Raw() float64 {
	return float64(*n)
}

func (n *Number) AsNumber() (*Number, error) {
	return n, nil
}

func (*Number) AsString() (*String, error) {
	return nil, InvalidCastError{}
}

func (*Number) AsFunction() (Function, error) {
	return nil, InvalidCastError{}
}
