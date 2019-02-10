package std

import (
	"errors"
	"fmt"

	"github.com/tobyjsullivan/chalk/resolver/engine/types"
)

var Sum = func(params []*types.Object) (*types.Object, error) {
	numbers := make([]float64, len(params))
	for i, p := range params {
		cur, err := p.ToNumber()
		if err != nil {
			return nil, errors.New(fmt.Sprintf("unexpected param type %d: %s", i, err))
		}

		numbers[i] = cur
	}

	return types.NewNumber(sum(numbers...)), nil
}

func sum(numbers ...float64) float64 {
	var acc float64
	for _, n := range numbers {
		acc += n
	}
	return acc
}
