package std

import (
	"errors"
	"fmt"
	"github.com/tobyjsullivan/chalk/functions"
	"github.com/tobyjsullivan/chalk/types"
)

var Sum = &functions.Function{
	Handler: func(params ...types.Object) (types.Object, error) {
		var acc float64
		for i, p := range params {
			cur, err := p.AsNumber()
			if err != nil {
				return nil, errors.New(fmt.Sprintf("unexpected param type %d", i))
			}

			acc += cur.Raw()
		}

		return types.NewNumber(acc), nil
	},
	Parameters: []types.Type{types.TNumber, types.TNumber},
	Variadic: true,
	Returns: types.TNumber,
}
