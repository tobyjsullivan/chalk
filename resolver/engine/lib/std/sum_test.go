package std

import (
	"testing"

	"github.com/tobyjsullivan/chalk/resolver/engine/types"
)

func TestSum_Handler(t *testing.T) {
	result, err := Sum([]*types.Object{
		types.NewNumber(14),
		types.NewNumber(66),
		types.NewNumber(55.6),
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	s, err := result.ToNumber()
	if err != nil {
		t.Errorf("Unexpected cast error: %v", err)
	}

	if raw := s; raw != 135.6 {
		t.Errorf("Unexpected result value: %f", raw)
	}
}
