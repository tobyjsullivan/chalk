package functions

import "github.com/tobyjsullivan/chalk/types"

type Function struct {
	Handler    FunctionHandler
	Parameters []types.Type
	Variadic   bool // Indicates the last parameter can be repeated any number of times.
	Returns    types.Type
}

type FunctionHandler func(params ...types.Object) (types.Object, error)
