package engine

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/tobyjsullivan/chalk/monolith"

	"github.com/tobyjsullivan/chalk/resolver"
	"github.com/tobyjsullivan/chalk/resolver/engine/lib/std"
	"github.com/tobyjsullivan/chalk/resolver/engine/parsing"
	"github.com/tobyjsullivan/chalk/resolver/engine/types"
)

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

func parseFormula(formula string) (*types.Object, error) {
	ast, err := parsing.Parse(formula)
	if err != nil {
		return nil, err
	}

	return mapAst(ast)
}

func mapAst(ast *parsing.ASTNode) (*types.Object, error) {
	if ast == nil {
		return nil, nil
	}
	if ast.NumberVal != nil {
		n, err := strconv.ParseFloat(*ast.NumberVal, 64)
		if err != nil {
			return nil, err
		}
		return types.NewNumber(n), nil
	} else if ast.StringVal != nil {
		return types.NewString(*ast.StringVal), nil
	} else if ast.TupleVal != nil {
		return nil, errors.New("tuple not handled")
	} else if ast.ListVal != nil {
		elements := ast.ListVal.Elements
		elObjs := make([]*types.Object, len(elements))
		var err error
		for i, e := range elements {
			elObjs[i], err = mapAst(e)
			if err != nil {
				return nil, err
			}
		}

		return types.NewList(elObjs), nil
	} else if ast.RecordVal != nil {
		// TODO
		return nil, errors.New("record not handled")
	} else if ast.FunctionCall != nil {
		args := make([]*types.Object, len(ast.FunctionCall.Argument.Elements))
		for i, arg := range ast.FunctionCall.Argument.Elements {
			var err error
			args[i], err = mapAst(arg)
			if err != nil {
				return nil, err
			}
		}

		return types.NewApplication(ast.FunctionCall.FuncName, args), nil
	} else if ast.VariableVal != nil {
		return types.NewVariable(*ast.VariableVal), nil
	} else {
		return nil, fmt.Errorf("unknown ast node: %v", ast)
	}
}

func (e *Engine) resolve(ctx context.Context, formula *types.Object, varHistory []string) (*types.Object, error) {
	if formula == nil {
		return nil, nil
	}

	switch formula.Type() {
	case types.TypeNumber:
		return formula, nil
	case types.TypeString:
		return formula, nil
	case types.TypeList:
		return formula, nil
	case types.TypeRecord:
		return formula, nil
	case types.TypeVariable:
		v, _ := formula.ToVariable()
		return e.resolveVariable(ctx, v, varHistory)
	case types.TypeApplication:
		a, _ := formula.ToApplication()
		return e.resolveApplication(ctx, a, varHistory)
	default:
		return nil, errors.New(fmt.Sprintf("unrecognized argument type %s", formula.Type()))
	}
}

func (e *Engine) resolveVariable(ctx context.Context, variable *types.Variable, varHistory []string) (*types.Object, error) {
	varName := variable.Name
	// Check for cycles
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

func (e *Engine) resolveApplication(ctx context.Context, app *types.Application, varHistory []string) (*types.Object, error) {
	f, err := findFunction(app.FunctionName)
	if err != nil {
		return nil, err
	}

	arguments := make([]*types.Object, len(app.Arguments))
	for i, arg := range app.Arguments {
		arguments[i] = arg

		// Lookup any variable values.
		if arg.Type() == types.TypeVariable {
			v, _ := arg.ToVariable()

			arguments[i], err = e.resolveVariable(ctx, v, varHistory)
			if err != nil {
				return nil, err
			}
		}
	}

	// Resolve any nested applications.
	for i, arg := range arguments {
		if arg.Type() == types.TypeApplication {
			a, _ := arg.ToApplication()
			arguments[i], err = e.resolveApplication(ctx, a, varHistory)
			if err != nil {
				return nil, err
			}
		}
	}

	return f(arguments)
}

func toResult(res *types.Object) *resolver.ResolveResponse {
	obj, err := toResultObject(res)
	if err != nil {
		return toErrorResult(err)
	}

	return &resolver.ResolveResponse{
		Result: obj,
	}
}

func toResultObject(obj *types.Object) (*resolver.Object, error) {
	if obj == nil {
		return nil, nil
	}
	switch obj.Type() {
	case types.TypeNumber:
		n, _ := obj.ToNumber()
		return &resolver.Object{
			Type:        resolver.ObjectType_NUMBER,
			NumberValue: n,
		}, nil
	case types.TypeString:
		s, _ := obj.ToString()
		return &resolver.Object{
			Type:        resolver.ObjectType_STRING,
			StringValue: s,
		}, nil
	case types.TypeList:
		list, _ := obj.ToList()

		els := make([]*resolver.Object, len(list.Elements))
		var err error
		for i, el := range list.Elements {
			els[i], err = toResultObject(el)
			if err != nil {
				return nil, err
			}
		}

		return &resolver.Object{
			Type: resolver.ObjectType_LIST,
			ListValue: &resolver.List{
				Elements: els,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected result type: %v", obj.Type())
	}
}

func toErrorResult(err error) *resolver.ResolveResponse {
	return &resolver.ResolveResponse{
		Error: fmt.Sprint(err),
	}
}

func findFunction(funcName string) (types.Function, error) {
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
