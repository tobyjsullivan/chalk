package parsing

import (
	"github.com/tobyjsullivan/chalk/api"
	"io"
	"strings"
)

const (
	tokenTypePunctuation TokenType = iota
	tokenTypeNumber TokenType = iota
	tokenTypeString TokenType = iota
	tokenTypeKeyword TokenType = iota
	tokenTypeOperator TokenType = iota
	tokenTypeIdentifier TokenType = iota
)

type TokenType int

func Parse(formula string) (*api.Object, error) {
	strings.NewReader(formula)
}

type InputStream struct {
	input []rune
	pos int
	line int
	col int
}

func NewInputStream(input string) *InputStream {
	return &InputStream{
		input: []rune(input),
		pos: 0,
		line: 1,
		col: 0,
	}
}

func (is *InputStream) next() rune {
	ch := is.input[is.pos]
	is.pos++
	if ch == '\n' {
		is.line++
		is.col = 0
	} else {
		is.col++
	}
	return ch
}

func (is *InputStream) peek() rune {
	return is.input[is.pos]
}

func (is *InputStream) eot() bool {
	return is.pos >= len(is.input)
}

// Reference: http://lisperator.net/pltut/parser/token-stream
type TokenStream struct {
	inputStream *InputStream
}

func (t *TokenStream) readNext() (Token, error) {

}

func isWhitespace(ch rune) bool {
	return strings.ContainsRune(" \t\n", ch)
}

func isDigit(ch rune) bool {
	return strings.ContainsRune("0123456789", ch)
}

func isDigit(ch rune) bool {
	return strings.ContainsRune("0123456789", ch)
}

type Token struct {
	Type int
	Value string
}