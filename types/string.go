package types

type String string

func NewString(text string) *String {
	s := String(text)
	return &s
}

func (s *String) Raw() string {
	return string(*s)
}

func (s *String) AsString() (*String, error) {
	return s, nil
}

func (s *String) AsNumber() (*Number, error) {
	return nil, InvalidCastError{}
}
