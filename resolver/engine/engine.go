package engine

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/tobyjsullivan/chalk/monolith"

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

func (e *Engine) Query(ctx context.Context, formula string) (*types.Object, error) {
	function, err := parseFormula(formula)
	if err != nil {
		return nil, err
	}

	return e.resolve(ctx, function, []string{})
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
	} else if ast.ApplicationVal != nil {
		exp, err := mapAst(ast.ApplicationVal.Expression)
		if err != nil {
			return nil, err
		}

		args := make([]*types.Object, len(ast.ApplicationVal.Argument.Elements))
		for i, arg := range ast.ApplicationVal.Argument.Elements {
			var err error
			args[i], err = mapAst(arg)
			if err != nil {
				return nil, err
			}
		}

		return types.NewApplication(exp, args), nil
	} else if ast.VariableVal != nil {
		return types.NewVariable(*ast.VariableVal), nil
	} else if ast.Lambda != nil {
		exp, err := mapAst(ast.Lambda.Expression)
		if err != nil {
			return nil, err
		}

		// We expect each element of the free vars tuple to be simple named variables.
		freeVars := make([]string, len(ast.Lambda.FreeVariables.Elements))
		for i, element := range ast.Lambda.FreeVariables.Elements {
			variable, err := mapAst(element)
			if err != nil {
				return nil, err
			}
			if t := variable.Type(); t != types.TypeVariable {
				return nil, fmt.Errorf("expected lambda param to be variable; found %+v", t)
			}

			v, _ := variable.ToVariable()
			freeVars[i] = v.Name
		}

		return types.NewLambda(freeVars, exp), nil
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
	case types.TypeFunction:
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
		Names: []string{varName},
	})
	if err != nil {
		return nil, err
	}

	var match *monolith.Variable
	for _, v := range resp.Values {
		if strings.ToLower(v.Name) == strings.ToLower(varName) {
			match = v
			break
		}
	}

	if match != nil {
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

	// Try to find a built-in value
	builtin := findBuiltinVariable(varName)
	if builtin != nil {
		return builtin, nil
	}

	if required {
		return nil, fmt.Errorf("variable `%s` is not defined", varName)
	}
	return nil, nil
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
	exp, err := e.resolve(ctx, app.Expression, varHistory)
	if err != nil {
		return nil, err
	}

	// Resolve all arguments
	resolvedArgs := make([]*types.Object, len(app.Arguments))
	for i, arg := range app.Arguments {
		resolvedArgs[i], err = e.resolve(ctx, arg, varHistory)
		if err != nil {
			return nil, err
		}
	}

	if exp.Type() == types.TypeFunction {
		// Execute functions inline.
		f, _ := exp.ToFunction()
		return f(resolvedArgs)
	}

	if exp.Type() != types.TypeLambda {
		return nil, fmt.Errorf("attempt to call non-callable: %s", exp.Type())
	}

	// Bind any arguments for lambdas
	l, _ := exp.ToLambda()
	varMap := make(map[string]*types.Object)
	for i, varName := range l.FreeVariables {
		if i >= len(resolvedArgs) {
			return nil, fmt.Errorf("incomplete var set provided. missing: %v", l.FreeVariables[i:])
		}
		varMap[varName] = resolvedArgs[i]
	}

	bound, err := bindVariables(l.Expression, varMap)
	if err != nil {
		return nil, err
	}

	return e.resolve(ctx, bound, varHistory)
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

		exp, err := bindVariables(a.Expression, varMap)
		if err != nil {
			return nil, err
		}

		return types.NewApplication(exp, args), nil
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
	case types.TypeFunction:
		return obj, nil
	default:
		return nil, fmt.Errorf("unexpected object type: %v", obj.Type())
	}
}

func findBuiltinVariable(varName string) *types.Object {
	switch strings.ToLower(varName) {
	case "sum":
		return types.NewFunction(std.Sum)
	case "concatenate":
		return types.NewFunction(std.Concatenate)
	case "love":
		return types.NewFunction(std.Love)
	case "list":
		return types.NewFunction(std.List)
	default:
		return nil
	}
}
