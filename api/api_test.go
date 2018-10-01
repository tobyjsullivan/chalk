package api

import "testing"

func TestQuery(t *testing.T) {
	req := &QueryRequest{
		Application: &Application{
			FunctionName: "SUM",
			Arguments: []*Argument{
				{Type: TypeNumber, NumberValue: 1.0},
				{Type: TypeNumber, NumberValue: 2.0},
				{Type: TypeNumber, NumberValue: 3.0},
			},
		},
	}

	res := Query(req)

	if res.Error != "" {
		t.Errorf("Unexpected error response: %s", res.Error)
	}

	if res.Result.Type != TypeNumber {
		t.Errorf("Unexpected result type: %s", res.Result.Type)
	}

	if v := res.Result.NumberValue; v != 6.0 {
		t.Errorf("Unexpected result value: %f", v)
	}
}

func TestQueryNested(t *testing.T) {
	innerApp := &Application{
		FunctionName: "CONCATENATE",
		Arguments: []*Argument{
			{Type: TypeString, StringValue: "World"},
			{Type: TypeString, StringValue: "!"},
		},
	}

	req := &QueryRequest{
		Application: &Application{
			FunctionName: "CONCATENATE",
			Arguments: []*Argument{
				{Type: TypeString, StringValue: "Hello, "},
				{Type: TypeApplication, ApplicationValue: innerApp},
			},
		},
	}

	res := Query(req)

	if res.Error != "" {
		t.Errorf("Unexpected error response: %s", res.Error)
	}

	if res.Result.Type != TypeString {
		t.Errorf("Unexpected result type: %s", res.Result.Type)
	}

	if v := res.Result.StringValue; v != "Hello, World!" {
		t.Errorf("Unexpected result value: %s", v)
	}
}
