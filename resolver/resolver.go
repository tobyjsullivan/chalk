package resolver

import (
	"errors"
	"fmt"
	"github.com/tobyjsullivan/chalk/functions"
	"github.com/tobyjsullivan/chalk/lib/std"
	"github.com/tobyjsullivan/chalk/parsing"
	"github.com/tobyjsullivan/chalk/resolver/rpc"
	"github.com/tobyjsullivan/chalk/types"
	"strconv"
	"strings"
)

const (
	typeString      = "string"
	typeNumber      = "number"
	typeApplication = "application"
)

type application struct {
	FunctionName string    `json:"function"`
	Arguments    []*object `json:"arguments"`
}

type object struct {
	Type             string       `json:"type"`
	StringValue      string       `json:"stringValue,omitempty"`
	NumberValue      float64      `json:"numberValue,omitempty"`
	ApplicationValue *application `json:"applicationValue,omitempty"`
}

func Query(request *rpc.ResolveRequest) *rpc.ResolveResponse {
	ast, err := parsing.Parse(request.Formula)
	if err != nil {
		return toErrorResult(err)
	}

	function, err := mapAst(ast)

	result, err := resolve(function)
	if err != nil {
		return toErrorResult(err)
	}

	return toResult(result)
}

func mapAst(ast *parsing.ASTNode) (*object, error) {
	if ast.NumberVal != nil {
		f, err := strconv.ParseFloat(*ast.NumberVal, 64)
		if err != nil {
			return nil, err
		}
		return &object{
			Type:        typeNumber,
			NumberValue: f,
		}, nil
	} else if ast.StringVal != nil {
		return &object{
			Type:        typeString,
			StringValue: *ast.StringVal,
		}, nil
	} else if ast.FunctionCall != nil {
		args := make([]*object, len(ast.FunctionCall.Arguments))
		for i, arg := range ast.FunctionCall.Arguments {
			var err error
			args[i], err = mapAst(arg)
			if err != nil {
				return nil, err
			}
		}

		return &object{
			Type: typeApplication,
			ApplicationValue: &application{
				FunctionName: ast.FunctionCall.FuncName,
				Arguments:    args,
			},
		}, nil
	} else {
		return nil, fmt.Errorf("unknown ast node: %v", ast)
	}
}

func isScalar(formula *object) (bool, error) {
	switch formula.Type {
	case typeApplication:
		return false, nil
	case typeNumber:
		return true, nil
	case typeString:
		return true, nil
	default:
		return false, errors.New(fmt.Sprintf("unrecognized argument type %s", formula.Type))
	}
}

func resolve(formula *object) (*object, error) {
	if isScalar, err := isScalar(formula); err != nil {
		return nil, err
	} else if isScalar {
		return formula, nil
	}

	app, err := toApplication(formula.ApplicationValue)
	if err != nil {
		return nil, err
	}
	result, err := app.Resolve()
	if err != nil {
		return nil, err
	}

	return fromFuncObject(result)
}

func toApplication(app *application) (*functions.Application, error) {
	f, err := findFunction(app.FunctionName)
	if err != nil {
		return nil, err
	}

	args := make([]functions.Argument, len(app.Arguments))
	for i, arg := range app.Arguments {
		args[i], err = toArgument(arg)
		if err != nil {
			return nil, err
		}
	}

	return &functions.Application{
		Function:  f,
		Arguments: args,
	}, nil
}

func fromFuncObject(input types.Object) (*object, error) {
	var output *object
	switch e := input.(type) {
	case *types.Number:
		output = &object{
			Type:        typeNumber,
			NumberValue: e.Raw(),
		}
	case *types.String:
		output = &object{
			Type:        typeString,
			StringValue: e.Raw(),
		}
	default:
		return nil, errors.New(fmt.Sprintf("unrecognized object type %v", input))
	}

	return output, nil
}

func toResult(res *object) *rpc.ResolveResponse {
	switch res.Type {
	case typeNumber:
		return &rpc.ResolveResponse{
			Result: &rpc.Object{
				Type:        rpc.ObjectType_NUMBER,
				NumberValue: res.NumberValue,
			},
		}
	case typeString:
		return &rpc.ResolveResponse{
			Result: &rpc.Object{
				Type:        rpc.ObjectType_STRING,
				StringValue: res.StringValue,
			},
		}
	default:
		return toErrorResult(fmt.Errorf("unexpected result type: %v", res.Type))
	}
}

func toErrorResult(err error) *rpc.ResolveResponse {
	return &rpc.ResolveResponse{
		Error: fmt.Sprint(err),
	}
}

func findFunction(funcName string) (*types.Function, error) {
	switch strings.ToLower(funcName) {
	case "sum":
		return std.Sum, nil
	case "concatenate":
		return std.Concatenate, nil
	default:
		return nil, errors.New("function not found")
	}
}

func toArgument(arg *object) (functions.Argument, error) {
	switch arg.Type {
	case typeApplication:
		app, err := toApplication(arg.ApplicationValue)
		if err != nil {
			return nil, err
		}
		return app, nil
	case typeNumber:
		return functions.NewArgument(types.NewNumber(arg.NumberValue)), nil
	case typeString:
		return functions.NewArgument(types.NewString(arg.StringValue)), nil
	default:
		return nil, errors.New(fmt.Sprintf("unrecognized argument type %s", arg.Type))
	}
}
