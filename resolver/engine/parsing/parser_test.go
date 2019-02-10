package parsing

import "testing"

func TestParser_ParseFunctions(t *testing.T) {
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

	if ast.FunctionCall.Argument == nil {
		t.Fatal("Expected arg tuple; got nil")
	}

	if n := len(ast.FunctionCall.Argument.Elements); n != 2 {
		t.Fatalf("Expected 2 elements in arg tuple, got %d", n)
	}

	if arg1 := ast.FunctionCall.Argument.Elements[0]; arg1 == nil || arg1.NumberVal == nil {
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
