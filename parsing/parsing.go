package parsing

import (
	"fmt"
	"github.com/tobyjsullivan/chalk/api"
	"strconv"
)

func Parse(formula string) (*api.Object, error) {
	p := NewParser(NewLexer(NewInputStream(formula)))

	ast, err := p.Parse()
	if err != nil {
		return nil, err
	}

	return mapAst(ast)
}

func mapAst(ast *ASTNode) (*api.Object, error) {
	if ast.NumberVal != nil {
		f, err := strconv.ParseFloat(*ast.NumberVal, 64)
		if err != nil {
			return nil, err
		}
		return &api.Object{
			Type: api.TypeNumber,
			NumberValue: f,
		}, nil
	} else if ast.StringVal != nil {
		return &api.Object{
			Type: api.TypeString,
			StringValue: *ast.StringVal,
		}, nil
	} else if ast.FunctionCall != nil {
		args := make([]*api.Object, len(ast.FunctionCall.Arguments))
		for i, arg := range ast.FunctionCall.Arguments {
			var err error
			args[i], err = mapAst(arg)
			if err != nil {
				return nil, err
			}
		}

		return &api.Object{
			Type:api.TypeApplication,
			ApplicationValue:&api.Application{
				FunctionName: ast.FunctionCall.FuncName,
				Arguments:args,
			},
		}, nil
	} else {
		return nil, fmt.Errorf("unknown ast node: %v", ast)
	}
}
