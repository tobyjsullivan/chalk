package engine

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/tobyjsullivan/chalk/monolith"

	"github.com/tobyjsullivan/chalk/resolver"
	"github.com/tobyjsullivan/chalk/resolver/functions"
	"github.com/tobyjsullivan/chalk/resolver/lib/std"
	"github.com/tobyjsullivan/chalk/resolver/parsing"
	"github.com/tobyjsullivan/chalk/resolver/types"
)

const (
	typeString      = "string"
	typeNumber      = "number"
	typeVariable    = "variable"
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
	VariableName     string       `json:"variableName,omitempty"`
	ApplicationValue *application `json:"applicationValue,omitempty"`
}

type Engine struct {
	varSvc monolith.VariablesClient
}

func NewEngine(varSvc monolith.VariablesClient) *Engine {
	return &Engine{
		varSvc: varSvc,
	}
}

func (e *Engine) Query(ctx context.Context, request *resolver.ResolveRequest) *resolver.ResolveResponse {
	function, err := parseFormula(request.Formula)
	if err != nil {
		return toErrorResult(err)
	}

	result, err := e.resolve(ctx, function, []string{})
	if err != nil {
		return toErrorResult(err)
	}

	return toResult(result)
}

func parseFormula(formula string) (*object, error) {
	ast, err := parsing.Parse(formula)
	if err != nil {
		return nil, err
	}

	return mapAst(ast)
}

func mapAst(ast *parsing.ASTNode) (*object, error) {
	if ast == nil {
		return nil, nil
	}
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
	} else if ast.VariableVal != nil {
		return &object{
			Type:         typeVariable,
			VariableName: *ast.VariableVal,
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
	case typeVariable:
		return false, nil
	default:
		return false, errors.New(fmt.Sprintf("unrecognized argument type %s", formula.Type))
	}
}

func (e *Engine) resolve(ctx context.Context, formula *object, varHistory []string) (*object, error) {
	if formula == nil {
		return nil, nil
	}

	if isScalar, err := isScalar(formula); err != nil {
		return nil, err
	} else if isScalar {
		return formula, nil
	}

	if formula.Type == typeVariable {
		// Check for cycles
		varName := formula.VariableName
		for _, seen := range varHistory {
			if seen == varName {
				return nil, fmt.Errorf("variable cycle detected: %s", varName)
			}
		}

		// Lookup formula
		resp, err := e.varSvc.GetVariables(ctx, &monolith.GetVariablesRequest{
			Keys: []string{varName},
		})
		if err != nil {
			return nil, err
		}
		if len(resp.Values) != 1 {
			return nil, fmt.Errorf("expected exactly 1 value; got: %d", len(resp.Values))
		}
		f := resp.Values[0].Formula

		// get object
		o, err := parseFormula(f)
		if err != nil {
			return nil, err
		}

		// resolve
		newHist := make([]string, len(varHistory)+1)
		copy(newHist, varHistory)
		newHist[len(varHistory)] = varName
		return e.resolve(ctx, o, newHist)
	}

	app, err := e.toApplication(ctx, formula.ApplicationValue, varHistory)
	if err != nil {
		return nil, err
	}
	result, err := app.Resolve()
	if err != nil {
		return nil, err
	}

	return fromFuncObject(result)
}

func (e *Engine) toApplication(ctx context.Context, app *application, varHistory []string) (*functions.Application, error) {
	f, err := findFunction(app.FunctionName)
	if err != nil {
		return nil, err
	}

	args := make([]functions.Argument, len(app.Arguments))
	for i, arg := range app.Arguments {
		r, err := e.resolve(ctx, arg, varHistory)
		if err != nil {
			return nil, err
		}
		args[i], err = toArgument(r)
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

func toResult(res *object) *resolver.ResolveResponse {
	if res == nil {
		return &resolver.ResolveResponse{
			Result: nil,
		}
	}
	switch res.Type {
	case typeNumber:
		return &resolver.ResolveResponse{
			Result: &resolver.Object{
				Type:        resolver.ObjectType_NUMBER,
				NumberValue: res.NumberValue,
			},
		}
	case typeString:
		return &resolver.ResolveResponse{
			Result: &resolver.Object{
				Type:        resolver.ObjectType_STRING,
				StringValue: res.StringValue,
			},
		}
	default:
		return toErrorResult(fmt.Errorf("unexpected result type: %v", res.Type))
	}
}

func toErrorResult(err error) *resolver.ResolveResponse {
	return &resolver.ResolveResponse{
		Error: fmt.Sprint(err),
	}
}

func findFunction(funcName string) (*types.Function, error) {
	switch strings.ToLower(funcName) {
	case "sum":
		return std.Sum, nil
	case "concatenate":
		return std.Concatenate, nil
	case "love":
		return std.Love, nil
	default:
		return nil, errors.New("function not found")
	}
}

func toArgument(arg *object) (functions.Argument, error) {
	if arg == nil {
		return nil, errors.New("unexpected nil argument passed to function")
	}

	switch arg.Type {
	case typeNumber:
		return functions.NewArgument(types.NewNumber(arg.NumberValue)), nil
	case typeString:
		return functions.NewArgument(types.NewString(arg.StringValue)), nil
	default:
		return nil, errors.New(fmt.Sprintf("unexpected argument type %s", arg.Type))
	}
}
