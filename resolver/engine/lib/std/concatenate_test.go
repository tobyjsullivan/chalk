package std

import (
	"testing"

	"github.com/tobyjsullivan/chalk/resolver/engine/types"
)

func TestConcatenate_Handler(t *testing.T) {
	result, err := Concatenate([]*types.Object{
		types.NewString("Hello"),
		types.NewString(", "),
		types.NewString("World!"),
	})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	s, err := result.ToString()
	if err != nil {
		t.Errorf("Unexpected cast error: %v", err)
	}

	if s != "Hello, World!" {
		t.Errorf("Unexpected result value: %s", s)
	}
}
