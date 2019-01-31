package types

import "testing"

func TestObject_AsString(t *testing.T) {
	var o Object = NewString("Hello, world.")

	s, err := o.AsString()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if raw := s.Raw(); raw != "Hello, world." {
		t.Errorf("Unexpected value: %s", raw)
	}
}
