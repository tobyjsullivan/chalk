package functions

import (
	"github.com/tobyjsullivan/chalk/types"
	"testing"
)

func TestApplication_Resolve(t *testing.T) {
	fSum := &Function{
		Handler: func(params ...types.Object) (types.Object, error) {
			a, _ := params[0].AsNumber()
			b, _ := params[1].AsNumber()

			return types.NewNumber(a.Raw() + b.Raw()), nil
		},
		Parameters: []types.Type{types.TNumber, types.TNumber},
		Returns: types.TNumber,
	}

	app := &Application{
		Function: fSum,
		Arguments: []Argument{
			NewArgument(types.NewNumber(32.5)),
			NewArgument(types.NewNumber(12.0)),
		},
	}

	result, err := app.Resolve()
	if err != nil {
		t.Errorf("Unexpected resolving error: %v", err)
	}

	nResult, err := result.AsNumber()
	if err != nil {
		t.Errorf("Unexpected cast error: %v", err)
	}

	if raw := nResult.Raw(); raw != 44.5 {
		t.Errorf("Unexpected value: %f", raw)
	}
}
