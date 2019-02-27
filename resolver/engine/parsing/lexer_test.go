package parsing

import "testing"

func TestLexer_ident(t *testing.T) {
	input := "SUM"

	lex := NewLexer(NewInputStream(input))

	if tok := lex.Next(); tok.Type != tokenIdentifier || tok.Value != "SUM" {
		t.Error("Expected ", tokenIdentifier, "SUM", "; got", tok.Type, tok.Value)
	}
}

func TestLexer_paren(t *testing.T) {
	input := "("

	lex := NewLexer(NewInputStream(input))

	if tok := lex.Next(); tok.Type != tokenPunctuation || tok.Value != "(" {
		t.Error("Expected ", tokenPunctuation, "(", "; got", tok.Type, tok.Value)
	}
}

func TestLexer_arrow(t *testing.T) {
	input := "=>"

	lex := NewLexer(NewInputStream(input))

	if tok := lex.Next(); tok.Type != tokenPunctuation || tok.Value != "=>" {
		t.Error("Expected ", tokenPunctuation, "=>", "; got", tok.Type, tok.Value)
	}
}

func TestLexer_integer(t *testing.T) {
	input := "42"

	lex := NewLexer(NewInputStream(input))

	if tok := lex.Next(); tok.Type != tokenNumber || tok.Value != "42" {
		t.Error("Expected ", tokenNumber, "42", "; got", tok.Type, tok.Value)
	}
}

func TestLexer_decimal(t *testing.T) {
	input := "39.234"

	lex := NewLexer(NewInputStream(input))

	expected := "39.234"
	if tok := lex.Next(); tok.Type != tokenNumber || tok.Value != expected {
		t.Error("Expected ", tokenNumber, expected, "; got", tok.Type, tok.Value)
	}
}

func TestLexer_keywordTrue(t *testing.T) {
	input := "True"
	lex := NewLexer(NewInputStream(input))

	if tok := lex.Next(); tok.Type != tokenKeyword || tok.Value != "True" {
		t.Errorf("Expected %d `True`; got: %d `%s`", tokenKeyword, tok.Type, tok.Value)
	}
}

func TestLexer_keywordFalse(t *testing.T) {
	input := "FALSE"
	lex := NewLexer(NewInputStream(input))

	if tok := lex.Next(); tok.Type != tokenKeyword || tok.Value != "FALSE" {
		t.Errorf("Expected %d `FALSE`; got: %d `%s`", tokenKeyword, tok.Type, tok.Value)
	}
}

func TestCurrying(t *testing.T) {
	input := "Var1()(b)"

	lex := NewLexer(NewInputStream(input))

	expected := "Var1"
	if tok := lex.Next(); tok.Type != tokenIdentifier || tok.Value != expected {
		t.Error("Expected ", tokenIdentifier, expected, "; got", tok.Type, tok.Value)
	}

	expected = "("
	if tok := lex.Next(); tok.Type != tokenPunctuation || tok.Value != expected {
		t.Error("Expected ", tokenPunctuation, expected, "; got", tok.Type, tok.Value)
	}

	expected = ")"
	if tok := lex.Next(); tok.Type != tokenPunctuation || tok.Value != expected {
		t.Error("Expected ", tokenPunctuation, expected, "; got", tok.Type, tok.Value)
	}

	expected = "("
	if tok := lex.Next(); tok.Type != tokenPunctuation || tok.Value != expected {
		t.Error("Expected ", tokenPunctuation, expected, "; got", tok.Type, tok.Value)
	}

	expected = "b"
	if tok := lex.Next(); tok.Type != tokenIdentifier || tok.Value != expected {
		t.Error("Expected ", tokenIdentifier, expected, "; got", tok.Type, tok.Value)
	}

	expected = ")"
	if tok := lex.Next(); tok.Type != tokenPunctuation || tok.Value != expected {
		t.Error("Expected ", tokenPunctuation, expected, "; got", tok.Type, tok.Value)
	}

	if tok := lex.Next(); tok != nil {
		t.Error("Expected EOF; got ", tok.Type, tok.Value)
	}
}

func TestLexer(t *testing.T) {
	input := "SUM(3, 6)"

	lex := NewLexer(NewInputStream(input))

	if tok := lex.Next(); tok.Type != tokenIdentifier || tok.Value != "SUM" {
		t.Error("Expected ", tokenIdentifier, "SUM", "; got", tok.Type, tok.Value)
	}

	if tok := lex.Next(); tok.Type != tokenPunctuation || tok.Value != "(" {
		t.Error("Expected ", tokenPunctuation, "(", "; got", tok.Type, tok.Value)
	}

	if tok := lex.Next(); tok.Type != tokenNumber || tok.Value != "3" {
		t.Error("Expected ", tokenNumber, "3", "; got", tok.Type, tok.Value)
	}

	if tok := lex.Next(); tok.Type != tokenPunctuation || tok.Value != "," {
		t.Error("Expected ", tokenPunctuation, ",", "; got", tok.Type, tok.Value)
	}

	if tok := lex.Next(); tok.Type != tokenNumber || tok.Value != "6" {
		t.Error("Expected ", tokenNumber, "6", "; got", tok.Type, tok.Value)
	}

	if tok := lex.Next(); tok.Type != tokenPunctuation || tok.Value != ")" {
		t.Error("Expected ", tokenPunctuation, ")", "; got", tok.Type, tok.Value)
	}

	if tok := lex.Next(); tok != nil {
		t.Error("Expected ", nil, "; got", tok.Type, tok.Value)
	}
}
