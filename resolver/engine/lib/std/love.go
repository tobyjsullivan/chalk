package std

import (
	"errors"
	"fmt"

	"github.com/tobyjsullivan/chalk/resolver/engine/types"
)

var Love = func(params []types.Object) (types.Object, error) {
	if l := len(params); l != 1 {
		return nil, fmt.Errorf("expected exactly one parameter; recieved: %d", l)
	}

	s, err := params[0].AsString()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("expected string, got: %s", err))
	}

	return types.NewString(love(s.Raw())), nil
}

func love(name string) string {
	return "I love you, " + name + "!"
}
