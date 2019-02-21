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
	tok := p.l.Peek()
	if tok == nil {
		return nil, nil
	}

	switch tok.Type {
	case tokenNumber:
		p.l.Next()
		return &ASTNode{
			NumberVal: &tok.Value,
		}, nil
	case tokenString:
		p.l.Next()
		return &ASTNode{
			StringVal: &tok.Value,
		}, nil
	case tokenPunctuation:
		switch tok.Value {
		case "{":
			rec, err := p.parseRecord()
			if err != nil {
				return nil, err
			}
			return &ASTNode{
				RecordVal: rec,
			}, nil
		case "[":
			list, err := p.parseList()
			if err != nil {
				return nil, err
			}
			return &ASTNode{
				ListVal: list,
			}, nil
		case "(":
			// Handle lambda expression
			lambda, err := p.parseLambda()
			if err != nil {
				return nil, err
			}
			return &ASTNode{
				Lambda: lambda,
			}, nil
		default:
			return nil, fmt.Errorf("expected Number, String, Identifier, `{`, or `[`; got: %+v", tok)
		}
	case tokenIdentifier:
		p.l.Next()
		fName := tok.Value
		next := p.l.Peek()
		if next != nil && next.Type == tokenPunctuation && next.Value == "(" {
			// Arguments tuple implies function call
			argTuple, err := p.parseTuple()
			if err != nil {
				return nil, err
			}

			return &ASTNode{
				FunctionCall: &FunctionCall{
					FuncName: fName,
					Argument: argTuple,
				},
			}, nil
		}

		// Must be a variable
		return &ASTNode{
			VariableVal: &fName,
		}, nil
	default:
		return nil, fmt.Errorf("expected Number, String, Identifier, `{`, or `[`; got: %+v", tok)
	}
}

func (p *Parser) parseList() (*List, error) {
	if open := p.l.Next(); open == nil || open.Type != tokenPunctuation || open.Value != "[" {
		return nil, fmt.Errorf("expected `[`; got %+v", open)
	}

	var elements []*ASTNode
	for {
		tok := p.l.Peek()
		if tok == nil {
			return nil, errors.New("unexpected end of input")
		}

		if tok.Type == tokenPunctuation && tok.Value == "]" {
			p.l.Next()
			break
		} else {
			el, err := p.parseEntity()
			if err != nil {
				return nil, err
			}
			elements = append(elements, el)

			// Expect comma or close.
			nTok := p.l.Next()
			if nTok == nil {
				return nil, errors.New("unexpected end of input")
			} else if nTok.Type == tokenPunctuation && nTok.Value == "]" {
				break
			} else if nTok.Type != tokenPunctuation || nTok.Value != "," {
				return nil, fmt.Errorf("expected `,` or `]`; got %+v", nTok)
			}
		}
	}

	return &List{
		Elements: elements,
	}, nil
}

func (p *Parser) parseRecord() (*Record, error) {
	if open := p.l.Next(); open == nil || open.Type != tokenPunctuation || open.Value != "{" {
		return nil, fmt.Errorf("expected `{`; got %+v", open)
	}

	var properties []*RecordProperty
	for {
		tok := p.l.Peek()
		if tok == nil {
			return nil, errors.New("unexpected end of input")
		}

		if tok.Type == tokenPunctuation && tok.Value == "}" {
			p.l.Next()
			break
		} else if tok.Type == tokenIdentifier {
			prop, err := p.parseRecordProperty()
			if err != nil {
				return nil, err
			}

			properties = append(properties, prop)

			// Expect comma or close.
			nTok := p.l.Next()
			if nTok == nil {
				return nil, errors.New("unexpected end of input")
			} else if nTok.Type == tokenPunctuation && nTok.Value == "}" {
				break
			} else if nTok.Type != tokenPunctuation || nTok.Value != "," {
				return nil, fmt.Errorf("expected `,` or `}`; got %+v", nTok)
			}
		} else {
			return nil, fmt.Errorf("expected identifier or `}`; got %+v", tok)
		}
	}

	return &Record{
		Properties: properties,
	}, nil
}

func (p *Parser) parseRecordProperty() (*RecordProperty, error) {
	tok := p.l.Next()
	if tok == nil {
		return nil, errors.New("unexpected end of input")
	}

	if tok.Type != tokenIdentifier {
		return nil, fmt.Errorf("expected identifier; got %+v", tok)
	}

	propName := tok.Value

	if eq := p.l.Next(); eq == nil || eq.Type != tokenPunctuation || eq.Value != "=" {
		return nil, fmt.Errorf("expected `=`; got %+v", eq)
	}

	propVal, err := p.parseEntity()
	if err != nil {
		return nil, err
	}

	return &RecordProperty{
		Name:  propName,
		Value: propVal,
	}, nil
}

func (p *Parser) parseTuple() (*Tuple, error) {
	if open := p.l.Next(); open == nil || open.Type != tokenPunctuation || open.Value != "(" {
		return nil, fmt.Errorf("expected `(`; got %+v", open)
	}

	var elements []*ASTNode
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

			elements = append(elements, arg)

			// Expect comma or close.
			nTok := p.l.Next()
			if nTok == nil {
				return nil, errors.New("unexpected end of input")
			} else if nTok.Type == tokenPunctuation && nTok.Value == ")" {
				break
			} else if nTok.Type != tokenPunctuation || nTok.Value != "," {
				return nil, fmt.Errorf("expected `,` or `)`; got %+v", nTok)
			}
		}
	}

	return &Tuple{
		Elements: elements,
	}, nil
}

func (p *Parser) parseLambda() (*Lambda, error) {
	freeVariables, err := p.parseFreeVariables()
	if err != nil {
		return nil, err
	}

	// Expect "=>"
	if arrow := p.l.Next(); arrow == nil || arrow.Type != tokenPunctuation || arrow.Value != "=>" {
		return nil, fmt.Errorf("expected `=>`; got %+v", arrow)
	}

	exp, err := p.parseEntity()
	if err != nil {
		return nil, err
	}

	return &Lambda{
		FreeVariables: freeVariables,
		Expression:    exp,
	}, nil
}

func (p *Parser) parseFreeVariables() ([]string, error) {
	if open := p.l.Next(); open == nil || open.Type != tokenPunctuation || open.Value != "(" {
		return nil, fmt.Errorf("expected `(`; got %+v", open)
	}

	var freeVars []string
	for {
		tok := p.l.Peek()
		if tok == nil {
			return nil, errors.New("unexpected end of input")
		} else if tok.Type == tokenPunctuation && tok.Value == ")" {
			p.l.Next()
			break
		} else if tok.Type == tokenIdentifier {
			v := p.l.Next()
			freeVars = append(freeVars, v.Value)

			// Expect comma or close.
			nTok := p.l.Next()
			if nTok == nil {
				return nil, errors.New("unexpected end of input")
			} else if nTok.Type == tokenPunctuation && nTok.Value == ")" {
				break
			} else if nTok.Type != tokenPunctuation || nTok.Value != "," {
				return nil, fmt.Errorf("expected `,` or `)`; got %+v", nTok)
			}
		} else {
			return nil, fmt.Errorf("expected identifier or `)`; got %+v", tok)
		}
	}

	return freeVars, nil
}

type ASTNode struct {
	FunctionCall *FunctionCall
	ListVal      *List
	NumberVal    *string
	RecordVal    *Record
	StringVal    *string
	TupleVal     *Tuple
	VariableVal  *string
	Lambda       *Lambda
}

type Lambda struct {
	FreeVariables []string
	Expression    *ASTNode
}

type FunctionCall struct {
	Argument *Tuple
	FuncName string
}

type List struct {
	Elements []*ASTNode
}

type Record struct {
	Properties []*RecordProperty
}

type RecordProperty struct {
	Name  string
	Value *ASTNode
}

type Tuple struct {
	Elements []*ASTNode
}
