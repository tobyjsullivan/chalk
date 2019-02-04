package std

import (
	"errors"
	"fmt"

	"github.com/tobyjsullivan/chalk/resolver/types"
)

var Concatenate = &types.Function{
	Parameters: []types.Type{types.TString, types.TString},
	Variadic:   true,
	Returns:    types.TString,
	Handler: func(params ...types.Object) (types.Object, error) {
		strings := make([]string, len(params))
		for i, p := range params {
			s, err := p.AsString()
			if err != nil {
				return nil, errors.New(fmt.Sprintf("unexpected param type %d: %s", i, err))
			}
			strings[i] = s.Raw()
		}

		return types.NewString(concatenate(strings...)), nil
	},
}

func concatenate(in ...string) string {
	var acc string
	for _, s := range in {
		acc += s
	}
	return acc
}
