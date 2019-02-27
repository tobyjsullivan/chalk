package std

import (
	"fmt"

	"github.com/tobyjsullivan/chalk/resolver/engine/types"
)

var If = func(params []*types.Object) (*types.Object, error) {
	if n := len(params); n != 3 {
		return nil, fmt.Errorf("expected exactly 3 parameters; found %d", n)
	}
	condition, err := params[0].ToBoolean()
	if err != nil {
		return nil, err
	}

	if condition {
		return params[1], nil
	} else {
		return params[2], nil
	}
}
