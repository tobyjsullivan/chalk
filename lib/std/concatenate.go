package std

import (
	"errors"
	"fmt"
	"github.com/tobyjsullivan/chalk/types"
)

var Concatenate = &types.Function{
	Parameters: []types.Type{types.TString, types.TString},
	Variadic:   true,
	Returns:    types.TString,
	Handler: func(params ...types.Object) (types.Object, error) {
		var acc string
		for i, p := range params {
			s, err := p.AsString()
			if err != nil {
				return nil, errors.New(fmt.Sprintf("unexpected param type %d: %s", i, err))
			}

			acc += s.Raw()
		}

		return types.NewString(acc), nil
	},
}
