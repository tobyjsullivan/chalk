package engine

import (
	"context"
	"testing"

	"github.com/tobyjsullivan/chalk/monolith"
	"google.golang.org/grpc"

	rpc "github.com/tobyjsullivan/chalk/resolver"
)

type fakeVarSvc struct {
}

func (*fakeVarSvc) GetVariables(ctx context.Context, in *monolith.GetVariablesRequest, opts ...grpc.CallOption) (*monolith.GetVariablesResponse, error) {
	return &monolith.GetVariablesResponse{
		Values: []*monolith.Variable{
			{
				Name:    "var1",
				Formula: "\"Hello\"",
			},
		},
	}, nil
}
func (*fakeVarSvc) SetVariable(ctx context.Context, in *monolith.SetVariableRequest, opts ...grpc.CallOption) (*monolith.SetVariableResponse, error) {
	return &monolith.SetVariableResponse{}, nil
}

func TestQuery(t *testing.T) {
	req := &rpc.ResolveRequest{
		Formula: "SUM(1, 2, 3)",
	}

	e := NewEngine(nil)
	res := e.Query(context.Background(), req)

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

	e := NewEngine(nil)
	res := e.Query(context.Background(), req)

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

func TestListWithVar(t *testing.T) {
	fakeVarSvc := &fakeVarSvc{}

	req := &rpc.ResolveRequest{
		Formula: "[var1]",
	}

	e := NewEngine(fakeVarSvc)
	res := e.Query(context.Background(), req)

	if res.Error != "" {
		t.Fatalf("Unexpected error response: %s", res.Error)
	}

	if res.Result.Type != rpc.ObjectType_LIST {
		t.Errorf("Unexpected result type: %s", res.Result.Type)
	}

	element := res.Result.ListValue.Elements[0]
	if element.Type != rpc.ObjectType_STRING {
		t.Errorf("Unexpected element type: %s", res.Result.Type)
	}

	if v := element.StringValue; v != "Hello" {
		t.Errorf("Unexpected element value: %s", v)
	}
}

func TestLambda(t *testing.T) {
	fakeVarSvc := &fakeVarSvc{}

	req := &rpc.ResolveRequest{
		Formula: "(a, b) => SUM(a, b)",
	}

	e := NewEngine(fakeVarSvc)
	res := e.Query(context.Background(), req)

	if res.Error != "" {
		t.Fatalf("Unexpected error response: %s", res.Error)
	}

	if res.Result.Type != rpc.ObjectType_LAMBDA {
		t.Errorf("Unexpected result type: %s", res.Result.Type)
	}

	if varCount := len(res.Result.LambdaValue.FreeVariables); varCount != 2 {
		t.Errorf("Unexpected free variable count: %d", varCount)
	}

	if varA := res.Result.LambdaValue.FreeVariables[0]; varA != "a" {
		t.Errorf("Unexpected free variable: %s", varA)
	}

	if varB := res.Result.LambdaValue.FreeVariables[1]; varB != "b" {
		t.Errorf("Unexpected free variable: %s", varB)
	}

}
