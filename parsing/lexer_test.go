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

	if tok := lex.Next(); tok.Type != tokenNumber || tok.Value != "39.234" {
		t.Error("Expected ", tokenNumber, "29.234", "; got", tok.Type, tok.Value)
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