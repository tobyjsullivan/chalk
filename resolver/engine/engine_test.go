package engine

import (
	"context"
	"testing"

	"github.com/tobyjsullivan/chalk/resolver/engine/types"

	"github.com/tobyjsullivan/chalk/monolith"
	"google.golang.org/grpc"
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
	fakeVarSvc := &fakeVarSvc{}
	req := "SUM(1, 2, 3)"

	e := NewEngine(fakeVarSvc)
	res, err := e.Query(context.Background(), req)

	if err != nil {
		t.Fatalf("Unexpected error response: %s", err)
	}

	n, err := res.ToNumber()
	if res.Type() != types.TypeNumber {
		t.Error("Unexpected cast error:", err)
	}
	if n != 6.0 {
		t.Errorf("Unexpected result value: %f", n)
	}
}

func TestQueryNested(t *testing.T) {
	fakeVarSvc := &fakeVarSvc{}
	req := "CONCATENATE(\"Hello, \", CONCATENATE(\"World\", \"!\"))"

	e := NewEngine(fakeVarSvc)
	res, err := e.Query(context.Background(), req)

	if err != nil {
		t.Fatalf("Unexpected error response: %s", err)
	}

	s, err := res.ToString()
	if err != nil {
		t.Error("Unexpected cast error:", err)
	}

	if s != "Hello, World!" {
		t.Errorf("Unexpected result value: %s", s)
	}
}

func TestListWithVar(t *testing.T) {
	fakeVarSvc := &fakeVarSvc{}
	req := "[var1]"

	e := NewEngine(fakeVarSvc)
	res, err := e.Query(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error response: %s", err)
	}

	l, err := res.ToList()
	if err != nil {
		t.Fatal("Unexpected cast error:", err)
	}
	if n := len(l.Elements); n != 1 {
		t.Errorf("Expected exactly 1 element; found %d", n)
	}

	s, err := l.Elements[0].ToString()
	if err != nil {
		t.Fatal("Unexpected cast error:", err)
	}
	if s != "Hello" {
		t.Errorf("Unexpected element value: %s", s)
	}
}

func TestLambda(t *testing.T) {
	fakeVarSvc := &fakeVarSvc{}

	req := "(a, b) => SUM(a, b)"

	e := NewEngine(fakeVarSvc)
	res, err := e.Query(context.Background(), req)

	if err != nil {
		t.Fatalf("Unexpected error response: %s", err)
	}

	l, err := res.ToLambda()
	if err != nil {
		t.Error("Unexpected error in cast:", err)
	}
	if n := len(l.FreeVariables); n != 2 {
		t.Errorf("Unexpected free variable count: %d", n)
	}
	if varA := l.FreeVariables[0]; varA != "a" {
		t.Errorf("Unexpected free variable: %s", varA)
	}
	if varB := l.FreeVariables[1]; varB != "b" {
		t.Errorf("Unexpected free variable: %s", varB)
	}
}

func TestBoolTrue(t *testing.T) {
	fakeVarSvc := &fakeVarSvc{}

	req := "TRUE"

	e := NewEngine(fakeVarSvc)
	res, err := e.Query(context.Background(), req)

	if err != nil {
		t.Fatalf("Unexpected error response: %s", err)
	}
	b, err := res.ToBoolean()
	if err != nil {
		t.Error("Unexpected error in cast:", err)
	}
	if !b {
		t.Error("Expected TRUE; got FALSE")
	}
}

func TestNumberNegative(t *testing.T) {
	fakeVarSvc := &fakeVarSvc{}

	req := "-34.9"

	e := NewEngine(fakeVarSvc)
	res, err := e.Query(context.Background(), req)

	if err != nil {
		t.Fatalf("Unexpected error response: %s", err)
	}
	n, err := res.ToNumber()
	if err != nil {
		t.Error("Unexpected error in cast:", err)
	}
	if n != -34.9 {
		t.Errorf("Expected -34.9; got %f", n)
	}
}
