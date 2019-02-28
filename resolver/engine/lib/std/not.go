package std

import (
	"fmt"

	"github.com/tobyjsullivan/chalk/resolver/engine/types"
)

var Not = func(params []*types.Object) (*types.Object, error) {
	if n := len(params); n != 1 {
		return nil, fmt.Errorf("expected exactly 1 parameter; found %d", n)
	}
	input, err := params[0].ToBoolean()
	if err != nil {
		return nil, err
	}

	return types.NewBoolean(!input), nil
}
