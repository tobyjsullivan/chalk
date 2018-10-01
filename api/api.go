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
	TypeString = "string"
	TypeNumber = "number"
	TypeApplication = "application"
)

type QueryRequest struct {
	Application *Application `json:"application"`
}

type Application struct {
	FunctionName string `json:"function"`
	Arguments []*Argument `json:"arguments"`
}

type Argument struct {
	Type string `json:"type"`
	StringValue string `json:"stringValue"`
	NumberValue float64 `json:"numberValue"`
	ApplicationValue *Application `json:"applicationValue"`
}

type QueryResult struct {
	Result *ResultObject `json:"result"`
	Error string `json:"error"`
}

type ResultObject struct {
	Type string `json:"type"`
	StringValue string `json:"stringValue,omit-empty"`
	NumberValue float64 `json:"numberValue,omit-empty"`
}

func Query(request *QueryRequest) *QueryResult {
	app, err := toApplication(request.Application)
	if err != nil {
		return toErrorResult(err)
	}
	result, err := app.Resolve()
	if err != nil {
		return toErrorResult(err)
	}
	return toResult(result)
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
		Function: f,
		Arguments: args,
	}, nil
}

func toResult(res types.Object) (*QueryResult) {
	var obj *ResultObject
	switch e := res.(type) {
	case *types.Number:
		obj = &ResultObject{
			Type: TypeNumber,
			NumberValue: e.Raw(),
		}
	case *types.String:
		obj = &ResultObject{
			Type: TypeString,
			StringValue: e.Raw(),
		}
	}

	return &QueryResult{
		Result: obj,
	}
}

func toErrorResult(err error) (*QueryResult) {
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

func toArgument(arg *Argument) (functions.Argument, error) {
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