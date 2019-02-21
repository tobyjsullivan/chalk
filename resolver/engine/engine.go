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
		props := make(map[string]*types.Object)

		var err error
		for _, prop := range ast.RecordVal.Properties {
			props[prop.Name], err = mapAst(prop.Value)
			if err != nil {
				return nil, err
			}
		}

		return types.NewRecord(props), nil
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
	} else if ast.Lambda != nil {
		exp, err := mapAst(ast.Lambda.Expression)
		if err != nil {
			return nil, err
		}
		return types.NewLambda(ast.Lambda.FreeVariables, exp), nil
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
		l, _ := formula.ToList()
		return e.resolveList(ctx, l, varHistory)
	case types.TypeRecord:
		r, _ := formula.ToRecord()
		return e.resolveRecord(ctx, r, varHistory)
	case types.TypeVariable:
		v, _ := formula.ToVariable()
		result, err := e.resolveVariable(ctx, v, varHistory, true)
		if err != nil {
			return nil, err
		}
		return result, nil
	case types.TypeApplication:
		a, _ := formula.ToApplication()
		return e.resolveApplication(ctx, a, varHistory)
	case types.TypeLambda:
		return formula, nil
	default:
		return nil, fmt.Errorf("unrecognized argument type %s", formula.Type())
	}
}

func (e *Engine) resolveVariable(ctx context.Context, variable *types.Variable, varHistory []string, required bool) (*types.Object, error) {
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

	var match *monolith.Variable
	for _, v := range resp.Values {
		if v.Name == varName {
			match = v
			break
		}
	}

	if match == nil {
		if required {
			return nil, fmt.Errorf("variable `%s` is not defined", varName)
		}
		return nil, nil
	}

	f := match.Formula

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

func (e *Engine) resolveList(ctx context.Context, list *types.List, varHistory []string) (*types.Object, error) {
	resolvedElements := make([]*types.Object, len(list.Elements))
	var err error
	for i, element := range list.Elements {
		resolvedElements[i], err = e.resolve(ctx, element, varHistory)
		if err != nil {
			return nil, err
		}
	}

	return types.NewList(resolvedElements), err
}

func (e *Engine) resolveRecord(ctx context.Context, rec *types.Record, varHistory []string) (*types.Object, error) {
	resolvedProps := make(map[string]*types.Object)

	var err error
	for key, value := range rec.Properties {
		resolvedProps[key], err = e.resolve(ctx, value, varHistory)
		if err != nil {
			return nil, err
		}
	}

	return types.NewRecord(resolvedProps), err
}

func (e *Engine) resolveApplication(ctx context.Context, app *types.Application, varHistory []string) (*types.Object, error) {
	f, err := e.findFunction(ctx, app.FunctionName, varHistory)
	if err != nil {
		return nil, err
	}

	arguments := make([]*types.Object, len(app.Arguments))
	for i, arg := range app.Arguments {
		arguments[i] = arg

		// Lookup any variable values.
		if arg.Type() == types.TypeVariable {
			v, _ := arg.ToVariable()

			arguments[i], err = e.resolveVariable(ctx, v, varHistory, true)
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

func (e *Engine) findFunction(ctx context.Context, funcName string, varHistory []string) (types.Function, error) {
	// Check if function is a defined variable.
	v, err := e.resolveVariable(ctx, &types.Variable{Name: funcName}, varHistory, false)
	if err != nil {
		return nil, err
	}
	if v != nil {
		if v.Type() != types.TypeLambda {
			return nil, fmt.Errorf("variable `%s` is not a lambda", funcName)
		}

		l, _ := v.ToLambda()
		return func(paramTuple []*types.Object) (*types.Object, error) {
			varMap := make(map[string]*types.Object)
			for i, varName := range l.FreeVariables {
				if i >= len(paramTuple) {
					return nil, fmt.Errorf("incomplete var set provided. missing: %v", l.FreeVariables[i:])
				}
				varMap[varName] = paramTuple[i]
			}

			bound, err := bindVariables(l.Expression, varMap)
			if err != nil {
				return nil, err
			}

			return e.resolve(ctx, bound, varHistory)
		}, nil
	}

	// Otherwise try a builtin
	return findBuiltinFunction(funcName)
}

func bindVariables(obj *types.Object, varMap map[string]*types.Object) (*types.Object, error) {
	switch obj.Type() {
	case types.TypeString:
		return obj, nil
	case types.TypeNumber:
		return obj, nil
	case types.TypeVariable:
		v, _ := obj.ToVariable()
		if value, ok := varMap[v.Name]; ok {
			return value, nil
		}
		return obj, nil
	case types.TypeList:
		l, _ := obj.ToList()
		elements := make([]*types.Object, len(l.Elements))
		var err error
		for i, v := range l.Elements {
			elements[i], err = bindVariables(v, varMap)
			if err != nil {
				return nil, err
			}
		}
		return types.NewList(elements), nil
	case types.TypeRecord:
		r, _ := obj.ToRecord()
		props := make(map[string]*types.Object)
		var err error
		for k, v := range r.Properties {
			props[k], err = bindVariables(v, varMap)
			if err != nil {
				return nil, err
			}
		}

		return types.NewRecord(props), nil
	case types.TypeApplication:
		a, _ := obj.ToApplication()
		args := make([]*types.Object, len(a.Arguments))
		var err error
		for i, arg := range a.Arguments {
			args[i], err = bindVariables(arg, varMap)
			if err != nil {
				return nil, err
			}
		}
		return types.NewApplication(a.FunctionName, args), nil
	case types.TypeLambda:
		l, _ := obj.ToLambda()

		// Create a copy of the varMap but without vars defined in this lambda. This allows "shadowing".
		subMap := make(map[string]*types.Object)
		for k, v := range varMap {
			subMap[k] = v
		}
		for _, v := range l.FreeVariables {
			delete(subMap, v)
		}
		exp, err := bindVariables(l.Expression, subMap)
		if err != nil {
			return nil, err
		}
		return types.NewLambda(l.FreeVariables, exp), nil
	default:
		return nil, fmt.Errorf("unexpected object type: %v", obj.Type())
	}
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
	case types.TypeRecord:
		record, _ := obj.ToRecord()

		props := make([]*resolver.RecordProperty, 0, len(record.Properties))
		for k, v := range record.Properties {
			value, err := toResultObject(v)
			if err != nil {
				return nil, err
			}

			props = append(props, &resolver.RecordProperty{
				Name:  k,
				Value: value,
			})
		}

		return &resolver.Object{
			Type: resolver.ObjectType_RECORD,
			RecordValue: &resolver.Record{
				Properties: props,
			},
		}, nil
	case types.TypeLambda:
		lambda, _ := obj.ToLambda()

		return &resolver.Object{
			Type: resolver.ObjectType_LAMBDA,
			LambdaValue: &resolver.Lambda{
				FreeVariables: lambda.FreeVariables,
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

func findBuiltinFunction(funcName string) (types.Function, error) {
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
