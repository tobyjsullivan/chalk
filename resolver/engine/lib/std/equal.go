package std

import (
	"errors"
	"fmt"

	"github.com/tobyjsullivan/chalk/resolver/engine/types"
)

var Equal = func(params []*types.Object) (*types.Object, error) {
	if n := len(params); n != 2 {
		return nil, fmt.Errorf("expected exactly 2 parameters; found %d", n)
	}

	res, err := compareObjects(params[0], params[1])
	if err != nil {
		return nil, err
	}
	return types.NewBoolean(res), nil
}

func compareObjects(left, right *types.Object) (bool, error) {
	// TODO(toby): Update when we support casting.
	if left.Type() != right.Type() {
		return false, nil
	}

	switch left.Type() {
	case types.TypeApplication:
		return compareApplications(left, right)
	case types.TypeBoolean:
		return compareBooleans(left, right)
	case types.TypeFunction:
		return compareFunctions(left, right)
	case types.TypeLambda:
		return compareLambdas(left, right)
	case types.TypeList:
		return compareLists(left, right)
	case types.TypeNumber:
		return compareNumbers(left, right)
	case types.TypeRecord:
		return compareRecords(left, right)
	case types.TypeString:
		return compareStrings(left, right)
	case types.TypeVariable:
		return compareVariables(left, right)
	default:
		return false, fmt.Errorf("unexpected type: %s", left.Type())
	}
}

func compareApplications(_, _ *types.Object) (bool, error) {
	return false, errors.New("unresolved applications cannot be compared")
}

func compareBooleans(l, r *types.Object) (bool, error) {
	left, err := l.ToBoolean()
	if err != nil {
		return false, err
	}

	right, err := r.ToBoolean()
	if err != nil {
		return false, err
	}

	return left == right, nil
}

func compareFunctions(_, _ *types.Object) (bool, error) {
	return false, errors.New("unresolved functions cannot be compared")
}

func compareLambdas(_, _ *types.Object) (bool, error) {
	return false, errors.New("unresolved lambdas cannot be compared")
}

func compareLists(l, r *types.Object) (bool, error) {
	left, err := l.ToList()
	if err != nil {
		return false, err
	}

	right, err := r.ToList()
	if err != nil {
		return false, err
	}

	if len(left.Elements) != len(right.Elements) {
		return false, nil
	}

	for i, l := range left.Elements {
		r := right.Elements[i]

		if res, err := compareObjects(l, r); err != nil || !res {
			return false, err
		}
	}

	return true, nil
}

func compareNumbers(l, r *types.Object) (bool, error) {
	left, err := l.ToNumber()
	if err != nil {
		return false, err
	}

	right, err := r.ToNumber()
	if err != nil {
		return false, err
	}

	return left == right, nil

}

func compareRecords(l, r *types.Object) (bool, error) {
	left, err := l.ToRecord()
	if err != nil {
		return false, err
	}

	right, err := r.ToRecord()
	if err != nil {
		return false, err
	}

	if len(left.Properties) != len(right.Properties) {
		return false, nil
	}

	for k, l := range left.Properties {
		r, ok := right.Properties[k]
		if !ok {
			// Right doesn't have this property
			return false, nil
		}

		if res, err := compareObjects(l, r); err != nil || !res {
			return false, err
		}
	}

	return true, nil
}

func compareStrings(l, r *types.Object) (bool, error) {
	left, err := l.ToString()
	if err != nil {
		return false, err
	}

	right, err := r.ToString()
	if err != nil {
		return false, err
	}

	return left == right, nil
}

func compareVariables(l, r *types.Object) (bool, error) {
	return false, errors.New("unresolved variables cannot be compared")
}
