package types

type Function struct {
	Handler    FunctionHandler
	Parameters []Type
	Variadic   bool // Indicates the last parameter can be repeated any number of times.
	Returns    Type
}

type FunctionHandler func(params ...Object) (Object, error)

func (f *Function) AsNumber() (*Number, error) {
	return nil, InvalidCastError{}
}

func (f *Function) AsString() (*String, error) {
	return nil, InvalidCastError{}
}

func (f *Function) AsFunction() (*Function, error) {
	return f, nil
}
