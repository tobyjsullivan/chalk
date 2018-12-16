package api

import (
	"errors"
	"fmt"
	"github.com/tobyjsullivan/chalk/functions"
	"github.com/tobyjsullivan/chalk/lib/std"
	"github.com/tobyjsullivan/chalk/types"
	"strings"
)

const (
	TypeString      = "string"
	TypeNumber      = "number"
	TypeApplication = "application"
)

type QueryRequest struct {
	Formula *Object `json:"formula"`
}

type Application struct {
	FunctionName string    `json:"function"`
	Arguments    []*Object `json:"arguments"`
}

type Object struct {
	Type             string       `json:"type"`
	StringValue      string       `json:"stringValue"`
	NumberValue      float64      `json:"numberValue"`
	ApplicationValue *Application `json:"applicationValue"`
}

type QueryResult struct {
	Result *Object `json:"result"`
	Error  string  `json:"error"`
}

func Query(request *QueryRequest) *QueryResult {
	result, err := resolve(request.Formula)
	if err != nil {
		return toErrorResult(err)
	}
	return toResult(result)
}

func isScalar(formula *Object) (bool, error) {
	switch formula.Type {
	case TypeApplication:
		return false, nil
	case TypeNumber:
		return true, nil
	case TypeString:
		return true, nil
	default:
		return false, errors.New(fmt.Sprintf("unrecognized argument type %s", formula.Type))
	}
}

func resolve(formula *Object) (*Object, error) {
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

func toApplication(app *Application) (*functions.Application, error) {
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

func fromFuncObject(input types.Object) (*Object, error) {
	var output *Object
	switch e := input.(type) {
	case *types.Number:
		output = &Object{
			Type:        TypeNumber,
			NumberValue: e.Raw(),
		}
	case *types.String:
		output = &Object{
			Type:        TypeString,
			StringValue: e.Raw(),
		}
	default:
		return nil, errors.New(fmt.Sprintf("unrecognized object type %v", input))
	}

	return output, nil
}

func toResult(res *Object) *QueryResult {
	return &QueryResult{
		Result: res,
	}
}

func toErrorResult(err error) *QueryResult {
	return &QueryResult{
		Error: fmt.Sprint(err),
	}
}

func findFunction(funcName string) (*functions.Function, error) {
	switch strings.ToLower(funcName) {
	case "sum":
		return std.Sum, nil
	case "concatenate":
		return std.Concatenate, nil
	default:
		return nil, errors.New("function not found")
	}
}

func toArgument(arg *Object) (functions.Argument, error) {
	switch arg.Type {
	case TypeApplication:
		app, err := toApplication(arg.ApplicationValue)
		if err != nil {
			return nil, err
		}
		return app, nil
	case TypeNumber:
		return functions.NewArgument(types.NewNumber(arg.NumberValue)), nil
	case TypeString:
		return functions.NewArgument(types.NewString(arg.StringValue)), nil
	default:
		return nil, errors.New(fmt.Sprintf("unrecognized argument type %s", arg.Type))
	}
}
