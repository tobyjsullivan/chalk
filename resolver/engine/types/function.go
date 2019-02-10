package types

type Function func(paramTuple []Object) (Object, error)

func (Function) AsNumber() (*Number, error) {
	return nil, InvalidCastError{}
}

func (Function) AsString() (*String, error) {
	return nil, InvalidCastError{}
}

func (f Function) AsFunction() (Function, error) {
	return f, nil
}
