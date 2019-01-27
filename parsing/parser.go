package parsing

import (
	"errors"
	"fmt"
)

// Reference: http://lisperator.net/pltut/parser/token-stream
type Parser struct {
	l *Lexer
}

func NewParser(l *Lexer) *Parser {
	return &Parser{
		l: l,
	}
}

func (p *Parser) Parse() (*ASTNode, error) {
	return p.parseEntity()
}

func (p *Parser) parseEntity() (*ASTNode, error) {
	tok := p.l.Next()
	switch tok.Type {
	case tokenNumber:
		return &ASTNode{
			NumberVal: &tok.Value,
		}, nil
	case tokenString:
		return &ASTNode{
			StringVal: &tok.Value,
		}, nil
	case tokenIdentifier:
		fName := tok.Value
		args, err := p.parseArguments()
		if err != nil {
			return nil, err
		}

		return &ASTNode{
			FunctionCall: &FunctionCall{
				FuncName:  fName,
				Arguments: args,
			},
		}, nil
	default:
		return nil, fmt.Errorf("expected Number, String, or Identifier; got: %v", tok)
	}
}

func (p *Parser) parseArguments() ([]*ASTNode, error) {
	if open := p.l.Next(); open == nil || open.Type != tokenPunctuation || open.Value != "(" {
		return nil, fmt.Errorf("expected `(`; got %v", open)
	}

	var args []*ASTNode
	for {
		tok := p.l.Peek()
		if tok == nil {
			return nil, errors.New("unexpected end of input")
		} else if tok.Type == tokenPunctuation && tok.Value == ")" {
			p.l.Next()
			break
		} else {
			arg, err := p.parseEntity()
			if err != nil {
				return nil, err
			}

			args = append(args, arg)

			// Expect comma or close.
			nTok := p.l.Next()
			if nTok == nil {
				return nil, errors.New("unexpected end of input")
			} else if nTok.Type == tokenPunctuation && nTok.Value == ")" {
				break
			} else if nTok.Type != tokenPunctuation || nTok.Value != "," {
				return nil, fmt.Errorf("expected `,` or `)`; got %v", nTok)
			}
		}
	}

	return args, nil
}

type ASTNode struct {
	StringVal    *string
	NumberVal    *string
	FunctionCall *FunctionCall
}

type FunctionCall struct {
	FuncName  string
	Arguments []*ASTNode
}
