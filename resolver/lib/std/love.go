package std

import (
	"errors"
	"fmt"
	"github.com/tobyjsullivan/chalk/resolver/types"
)

var Love = &types.Function{
	Parameters: []types.Type{types.TString},
	Variadic:   false,
	Returns:    types.TString,
	Handler: func(params ...types.Object) (types.Object, error) {
		if l := len(params); l != 1 {
			return nil, fmt.Errorf("expected exactly one parameter; recieved: %d", l)
		}

		s, err := params[0].AsString()
		if err != nil {
			return nil, errors.New(fmt.Sprintf("expected string, got: %s", err))
		}

		out := "I love you " + s.Raw() + "!"

		return types.NewString(out), nil
	},
}
