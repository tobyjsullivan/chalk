package parsing

import "testing"

func TestParser_ParseFunctions(t *testing.T) {
	input := "SUM(4, PRODUCT(3, 2))"
	p := NewParser(NewLexer(NewInputStream(input)))
	ast, err := p.Parse()

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if ast.ApplicationVal == nil {
		t.Fatalf("Expected SUM function; got: %v", ast)
	}

	if funcName := *ast.ApplicationVal.Expression.VariableVal; funcName != "SUM" {
		t.Errorf("Expected variable \"SUM\"; got %v", funcName)
	}

	if ast.ApplicationVal.Argument == nil {
		t.Fatal("Expected arg tuple; got nil")
	}

	if n := len(ast.ApplicationVal.Argument.Elements); n != 2 {
		t.Fatalf("Expected 2 elements in arg tuple, got %d", n)
	}

	if arg1 := ast.ApplicationVal.Argument.Elements[0]; arg1 == nil || arg1.NumberVal == nil {
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

func TestParse_Variable(t *testing.T) {
	input := "var1"

	p := NewParser(NewLexer(NewInputStream(input)))
	ast, err := p.Parse()

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if ast == nil {
		t.Fatalf("Expected ast, got nil")
	}

	if val := *ast.VariableVal; val != "var1" {
		t.Errorf("Expected `var1`; got %s", val)
	}
}

func TestParse_List(t *testing.T) {
	input := "[1, 2]"

	p := NewParser(NewLexer(NewInputStream(input)))
	ast, err := p.Parse()

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if ast.ListVal == nil {
		t.Fatalf("Expected list; got %+v", ast)
	}

	if n := len(ast.ListVal.Elements); n != 2 {
		t.Fatalf("Expected 2 elements; got %d", n)
	}

	if first := ast.ListVal.Elements[0]; first.NumberVal == nil || *first.NumberVal != "1" {
		t.Errorf("Expected number \"1\"; got %+v", first)
	}

	if second := ast.ListVal.Elements[1]; second.NumberVal == nil || *second.NumberVal != "2" {
		t.Errorf("Expected number \"2\"; got %+v", second)
	}
}

func TestParse_Record(t *testing.T) {
	input := "{ name=\"Jane Doe\", age=27 }"

	p := NewParser(NewLexer(NewInputStream(input)))
	ast, err := p.Parse()

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	rec := ast.RecordVal
	if n := len(rec.Properties); n != 2 {
		t.Fatalf("Expected 2 properties; got %d", n)
	}

	nameProp := rec.Properties[0]
	if nameProp.Name != "name" {
		t.Errorf("Expected property \"name\"; got %s", nameProp.Name)
	}
	if nameVal := *nameProp.Value.StringVal; nameVal != "Jane Doe" {
		t.Errorf("Expected value \"Jane Doe\"; got %s", nameVal)
	}

	ageProp := rec.Properties[1]
	if ageProp.Name != "age" {
		t.Errorf("Expected property \"age\"; got %s", ageProp.Name)
	}
	if ageVal := *ageProp.Value.NumberVal; ageVal != "27" {
		t.Errorf("Expected value \"27\"; got %s", ageVal)
	}
}

func TestParser_Lambda(t *testing.T) {
	input := "(a) => a"
	p := NewParser(NewLexer(NewInputStream(input)))
	ast, err := p.Parse()

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if ast.LambdaVal == nil {
		t.Fatalf("Expected lambda; got: %v", ast)
	}

	if n := len(ast.LambdaVal.FreeVariables.Elements); n != 1 {
		t.Errorf("Expected one argument; got %d", n)
	}

	if ast.LambdaVal.Expression.VariableVal == nil {
		t.Errorf("Expected variable; got %+v", ast.LambdaVal.Expression)
	}
}

func TestParser_BoolTrue(t *testing.T) {
	input := "true"
	p := NewParser(NewLexer(NewInputStream(input)))
	ast, err := p.Parse()

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if ast.BooleanVal == nil {
		t.Fatalf("Expected boolean; got: %v", ast)
	}

	if b := *ast.BooleanVal; !b {
		t.Errorf("Expected `true`; got %v", b)
	}
}

func TestParser_BoolFalse(t *testing.T) {
	input := "False"
	p := NewParser(NewLexer(NewInputStream(input)))
	ast, err := p.Parse()

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if ast.BooleanVal == nil {
		t.Fatalf("Expected boolean; got: %v", ast)
	}

	if b := *ast.BooleanVal; b {
		t.Errorf("Expected `false`; got %v", b)
	}
}

func TestParser_Currying(t *testing.T) {
	input := "var1()(a)"
	p := NewParser(NewLexer(NewInputStream(input)))
	ast, err := p.Parse()

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	outter := ast.ApplicationVal
	if outter == nil {
		t.Fatalf("Expected SUM function; got: %v", ast)
	}

	if outter.Expression.ApplicationVal == nil {
		t.Fatal("Expected an application")
	}

	inner := outter.Expression.ApplicationVal
	if inner == nil {
		t.Fatal("Expected an inner application")
	}

	if inner.Expression.VariableVal == nil {
		t.Fatalf("Expected a variable, var1; got: %+v", inner.Expression)
	}

	if vName := *inner.Expression.VariableVal; vName != "var1" {
		t.Errorf("Expected variable to be `var1`; got %s", vName)
	}

	if n := len(inner.Argument.Elements); n != 0 {
		t.Errorf("Expected 0 args on inner expression; got %d", n)
	}

	if n := len(outter.Argument.Elements); n != 1 {
		t.Errorf("Expected 1 arg on outter expression; got %d", n)
	}

	if arg := outter.Argument.Elements[0]; arg.VariableVal == nil {
		t.Errorf("Expected arg to be variable; got %+v", arg)
	}
}
