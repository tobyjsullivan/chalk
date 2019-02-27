package parsing

import (
	"errors"
	"fmt"
	"strings"
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
	res, err := p.parseEntity()
	if err != nil {
		return nil, err
	}

	// Expect EOF
	if next := p.l.Next(); next != nil {
		return nil, fmt.Errorf("expected EOF; found %+v", next)
	}

	return res, nil
}

func (p *Parser) parseEntity() (*ASTNode, error) {
	entity, err := p.parseImmediateEntity()
	if err != nil {
		return nil, err
	}

	// Check for a tuple argument set, indicating an expression.
	return p.maybeParseApplication(entity)
}

func (p *Parser) maybeParseApplication(entity *ASTNode) (*ASTNode, error) {
	n := p.l.Peek()
	if n == nil || n.Type != tokenPunctuation || n.Value != "(" {
		return entity, nil
	}

	t, err := p.parseTuple()
	if err != nil {
		return nil, err
	}

	return p.maybeParseApplication(&ASTNode{
		ApplicationVal: &Application{
			Expression: entity,
			Argument:   t,
		},
	})
}

func (p *Parser) parseImmediateEntity() (*ASTNode, error) {
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
			// Parse a tuple
			t, err := p.parseTuple()
			if err != nil {
				return nil, err
			}

			// Check if maybe a lambda?
			next := p.l.Peek()
			if next != nil && next.Type == tokenPunctuation && next.Value == "=>" {
				// Expect "=>"
				if arrow := p.l.Next(); arrow == nil || arrow.Type != tokenPunctuation || arrow.Value != "=>" {
					return nil, fmt.Errorf("expected `=>`; got %+v", arrow)
				}

				exp, err := p.parseEntity()
				if err != nil {
					return nil, err
				}

				return &ASTNode{
					LambdaVal: &Lambda{
						FreeVariables: t,
						Expression:    exp,
					},
				}, nil
			}

			// It's just a tuple
			return &ASTNode{
				TupleVal: t,
			}, nil
		default:
			return nil, fmt.Errorf("expected Number, String, Identifier, `{`, or `[`; got: %+v", tok)
		}
	case tokenKeyword:
		p.l.Next()
		switch strings.ToLower(tok.Value) {
		case "true":
			b := true
			return &ASTNode{
				BooleanVal: &b,
			}, nil
		case "false":
			b := false
			return &ASTNode{
				BooleanVal: &b,
			}, nil
		default:
			return nil, fmt.Errorf("unexpected keyword: `%s`", tok.Value)
		}
	case tokenIdentifier:
		p.l.Next()
		fName := tok.Value

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

type ASTNode struct {
	ApplicationVal *Application
	BooleanVal     *bool
	LambdaVal      *Lambda
	ListVal        *List
	NumberVal      *string
	RecordVal      *Record
	StringVal      *string
	TupleVal       *Tuple
	VariableVal    *string
}

func (n *ASTNode) String() string {
	if n.ApplicationVal != nil {
		return fmt.Sprintf(
			"Application{ Expression:%v, Argument:%v }",
			n.ApplicationVal.Expression,
			n.ApplicationVal.Argument,
		)
	}
	if n.ListVal != nil {
		return fmt.Sprintf("List{Elements:%v}", n.ListVal.Elements)
	}
	if n.NumberVal != nil {
		return fmt.Sprintf("Number{%v}", n.NumberVal)
	}
	if n.RecordVal != nil {
		return fmt.Sprintf("Record{%v}", n.RecordVal.Properties)
	}
	if n.StringVal != nil {
		return fmt.Sprintf("String{%v}", n.StringVal)
	}
	if n.TupleVal != nil {
		return fmt.Sprintf("Tuple{%v}", n.TupleVal.Elements)
	}
	if n.VariableVal != nil {
		return fmt.Sprintf("Variable{%v}", n.VariableVal)
	}
	if n.LambdaVal != nil {
		return fmt.Sprintf(
			"Lambda{FreeVariables:%v, Expression: %v}",
			n.LambdaVal.FreeVariables,
			n.LambdaVal.Expression,
		)
	}

	return "EmptyAST"
}

type Lambda struct {
	FreeVariables *Tuple
	Expression    *ASTNode
}

type Application struct {
	Expression *ASTNode
	Argument   *Tuple
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
