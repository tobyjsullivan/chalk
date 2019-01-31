package parsing

import "testing"

func TestParser_Parse(t *testing.T) {
	input := "SUM(4, PRODUCT(3, 2))"
	p := NewParser(NewLexer(NewInputStream(input)))
	ast, err := p.Parse()

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if ast.FunctionCall == nil {
		t.Fatalf("Expected SUM function; got: %v", ast)
	}

	if ast.FunctionCall.FuncName != "SUM" {
		t.Errorf("Expected function name \"SUM\"; got %v", ast.FunctionCall.FuncName)
	}

	if numArgs := len(ast.FunctionCall.Arguments); numArgs != 2 {
		t.Fatalf("Expected 2 args; got %d", numArgs)
	}

	if arg1 := ast.FunctionCall.Arguments[0]; arg1 == nil || arg1.NumberVal == nil {
		t.Fatalf("Expected number, got %v", arg1)
	} else if *arg1.NumberVal != "4" {
		t.Errorf("Expected \"4\"; got %s", *arg1.NumberVal)
	}
}

func TestParse_Empty(t *testing.T) {
	var input string

	p := NewParser(NewLexer(NewInputStream(input)))
	ast, err := p.Parse()

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if ast != nil {
		t.Fatalf("Expected ast to be nil, got %+v", ast)
	}
}
