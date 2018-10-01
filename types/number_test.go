package types

import "testing"

func TestObject_AsNumber(t *testing.T) {
	var o Object = NewNumber(3.145)

	n, err := o.AsNumber()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if float64(*n) != 3.145 {
		t.Errorf("Unexected value: %f", float64(*n))
	}
}
