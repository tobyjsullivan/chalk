package std

import (
	"errors"
	"fmt"

	"github.com/tobyjsullivan/chalk/resolver/engine/types"
)

var Concatenate = func(input []*types.Object) (*types.Object, error) {
	strings := make([]string, len(input))
	var err error
	for i, p := range input {
		strings[i], err = p.ToString()
		if err != nil {
			return nil, errors.New(fmt.Sprintf("unexpected param type #%d: %s", i, err))
		}
	}

	return types.NewString(concatenate(strings...)), nil
}

func concatenate(in ...string) string {
	var acc string
	for _, s := range in {
		acc += s
	}
	return acc
}
