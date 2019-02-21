package std

import (
	"github.com/tobyjsullivan/chalk/resolver/engine/types"
)

var List = func(input []*types.Object) (*types.Object, error) {
	elements := make([]*types.Object, len(input))
	copy(elements, input)

	return types.NewList(elements), nil
}
