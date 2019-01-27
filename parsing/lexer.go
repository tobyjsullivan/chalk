package parsing

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	tokenPunctuation TokenType = iota
	tokenNumber
	tokenString
	//tokenOperator
	tokenIdentifier
	tokenInvalid
)

var reIdentStart = regexp.MustCompile("^[a-zA-Z_]")

type TokenType int

// Source: http://lisperator.net/pltut/parser/token-stream
type Lexer struct {
	input   *InputStream
	pos     int
	current *Token
}

func NewLexer(input *InputStream) *Lexer {
	return &Lexer{
		input: input,
		pos:   0,
	}
}

func (l *Lexer) readNext() *Token {
	// Consume and ignore whitespace
	l.readWhile(func(r rune) bool {
		return isWhitespace(r)
	})
	if l.input.eof() {
		return nil
	}
	ch := l.input.peek()
	l.pos++
	if ch == '"' {
		return l.readString()
	}
	if isDigit(ch) {
		return l.readNumber()
	}
	if isPunctuation(ch) {
		return &Token{
			Type:  tokenPunctuation,
			Value: string(l.input.next()),
		}
	}
	if isIdentStart(ch) {
		return l.readIdentifier()
	}

	return &Token{
		Type:  tokenInvalid,
		Value: string(l.input.next()),
	}
}

func (l *Lexer) readWhile(p func(rune) bool) string {
	var str []rune
	for !l.input.eof() && p(l.input.peek()) {
		str = append(str, l.input.next())
	}

	return string(str)
}

func (l *Lexer) readString() *Token {
	var escaped bool
	var str []rune
	end := l.input.next()
	for {
		if l.input.eof() {
			// String wasn't closed
			return &Token{
				Type:  tokenIdentifier,
				Value: fmt.Sprint(end, string(str)),
			}
		}

		ch := l.input.next()
		if escaped {
			str = append(str, ch)
			escaped = false
		} else if ch == '\\' {
			escaped = true
		} else if ch == end {
			break
		} else {
			str = append(str, ch)
		}
	}

	return &Token{
		Type:  tokenString,
		Value: string(str),
	}
}

func (l *Lexer) readNumber() *Token {
	var decimal bool
	var str []rune
	for !l.input.eof() {
		ch := l.input.peek()
		if isDigit(ch) {
			str = append(str, l.input.next())
		} else if ch == '.' && !decimal {
			str = append(str, l.input.next())
			decimal = true
		} else {
			break
		}
	}

	return &Token{
		Type:  tokenNumber,
		Value: string(str),
	}
}

func (l *Lexer) readIdentifier() *Token {
	var str []rune
	for !l.input.eof() {
		ch := l.input.peek()
		if isIdent(ch) {
			str = append(str, l.input.next())
		} else {
			break
		}
	}

	return &Token{
		Type:  tokenIdentifier,
		Value: string(str),
	}
}

func (l *Lexer) Next() *Token {
	if l.current != nil {
		tmp := l.current
		l.current = nil
		return tmp
	}

	return l.readNext()
}

func (l *Lexer) Peek() *Token {
	if l.current != nil {
		return l.current
	}

	l.current = l.readNext()
	return l.current
}

func (l *Lexer) eof() bool {
	return l.Peek() == nil
}

func isWhitespace(ch rune) bool {
	return strings.ContainsRune(" \t\n", ch)
}

func isDigit(ch rune) bool {
	return strings.ContainsRune("0123456789", ch)
}

func isPunctuation(ch rune) bool {
	return strings.ContainsRune("(),", ch)
}

func isIdentStart(ch rune) bool {
	return reIdentStart.MatchString(string(ch))
}

func isIdent(ch rune) bool {
	return isIdentStart(ch) || strings.Index("0123456789", string(ch)) >= 0
}

type Token struct {
	Type  TokenType
	Value string
}
