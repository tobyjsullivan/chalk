package std

import (
	"github.com/tobyjsullivan/chalk/resolver/engine/types"
	"testing"
)

func TestSum_Handler(t *testing.T) {
	result, err := Sum([]types.Object{
		types.NewNumber(14),
		types.NewNumber(66),
		types.NewNumber(55.6),
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	s, err := result.AsNumber()
	if err != nil {
		t.Errorf("Unexpected cast error: %v", err)
	}

	if raw := s.Raw(); raw != 135.6 {
		t.Errorf("Unexpected result value: %f", raw)
	}
}
