package engine

import (
	"testing"

	rpc "github.com/tobyjsullivan/chalk/resolver"
)

func TestQuery(t *testing.T) {
	req := &rpc.ResolveRequest{
		Formula: "SUM(1, 2, 3)",
	}

	res := Query(req)

	if res.Error != "" {
		t.Fatalf("Unexpected error response: %s", res.Error)
	}

	if res.Result.Type != rpc.ObjectType_NUMBER {
		t.Errorf("Unexpected result type: %s", res.Result.Type)
	}

	if v := res.Result.NumberValue; v != 6.0 {
		t.Errorf("Unexpected result value: %f", v)
	}
}

func TestQueryNested(t *testing.T) {
	req := &rpc.ResolveRequest{
		Formula: "CONCATENATE(\"Hello, \", CONCATENATE(\"World\", \"!\"))",
	}

	res := Query(req)

	if res.Error != "" {
		t.Fatalf("Unexpected error response: %s", res.Error)
	}

	if res.Result.Type != rpc.ObjectType_STRING {
		t.Errorf("Unexpected result type: %s", res.Result.Type)
	}

	if v := res.Result.StringValue; v != "Hello, World!" {
		t.Errorf("Unexpected result value: %s", v)
	}
}
