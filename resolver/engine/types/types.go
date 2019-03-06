package types

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	TypeApplication TypeName = "application"
	TypeBoolean              = "boolean"
	TypeFunction             = "function" // A function differs from a lambda in that it executes code to resolve.
	TypeList                 = "list"
	TypeNumber               = "number"
	TypeLambda               = "lambda"
	TypeRecord               = "record"
	TypeString               = "string"
	TypeVariable             = "variable"
)

type TypeName string

type Application struct {
	Expression *Object
	Arguments  []*Object
}

type List struct {
	Elements []*Object
}

type Lambda struct {
	FreeVariables []string
	Expression    *Object
}

type Record struct {
	Properties map[string]*Object
}

type Variable struct {
	Name string
}

type Object struct {
	objectType       TypeName
	applicationValue *Application
	booleanValue     bool
	functionValue    Function
	listValue        *List
	numberValue      float64
	lambdaValue      *Lambda
	recordValue      *Record
	stringValue      string
	variableValue    *Variable
}

func NewString(s string) *Object {
	return &Object{
		objectType:  TypeString,
		stringValue: s,
	}
}

func NewNumber(n float64) *Object {
	return &Object{
		objectType:  TypeNumber,
		numberValue: n,
	}
}

func NewBoolean(b bool) *Object {
	return &Object{
		objectType:   TypeBoolean,
		booleanValue: b,
	}
}

func NewApplication(expression *Object, args []*Object) *Object {
	return &Object{
		objectType: TypeApplication,
		applicationValue: &Application{
			Expression: expression,
			Arguments:  args,
		},
	}
}

func NewFunction(f Function) *Object {
	return &Object{
		objectType:    TypeFunction,
		functionValue: f,
	}
}

func NewLambda(freeVariables []string, expression *Object) *Object {
	return &Object{
		objectType: TypeLambda,
		lambdaValue: &Lambda{
			FreeVariables: freeVariables,
			Expression:    expression,
		},
	}
}

func NewList(elements []*Object) *Object {
	return &Object{
		objectType: TypeList,
		listValue: &List{
			Elements: elements,
		},
	}
}

func NewRecord(properties map[string]*Object) *Object {
	return &Object{
		objectType: TypeRecord,
		recordValue: &Record{
			Properties: properties,
		},
	}
}

func NewVariable(varName string) *Object {
	return &Object{
		objectType: TypeVariable,
		variableValue: &Variable{
			Name: varName,
		},
	}
}

func (o *Object) Type() TypeName {
	return o.objectType
}

func (o *Object) ToString() (string, error) {
	if o.objectType == TypeString {
		return o.stringValue, nil
	}

	// Casting
	if o.objectType == TypeNumber {
		return strconv.FormatFloat(o.numberValue, 'f', -1, 64), nil
	}

	return "", fmt.Errorf("value is not a string: %+v", o)
}

func (o *Object) ToNumber() (float64, error) {
	if o.objectType == TypeNumber {
		return o.numberValue, nil
	}

	// Casting
	if o.objectType == TypeString {
		return strconv.ParseFloat(o.stringValue, 64)
	}

	return 0, errors.New("value is not a number")
}

func (o *Object) ToApplication() (*Application, error) {
	if o.objectType != TypeApplication {
		return nil, errors.New("value is not an application")
	}

	return o.applicationValue, nil
}

func (o *Object) ToBoolean() (bool, error) {
	if o.objectType != TypeBoolean {
		return false, errors.New("value is not an application")
	}

	return o.booleanValue, nil
}

func (o *Object) ToFunction() (Function, error) {
	if o.objectType != TypeFunction {
		return nil, errors.New("value is not a function")
	}

	return o.functionValue, nil
}

func (o *Object) ToLambda() (*Lambda, error) {
	if o.objectType != TypeLambda {
		return nil, errors.New("value is not a lambda")
	}

	return o.lambdaValue, nil
}

func (o *Object) ToList() (*List, error) {
	if o.objectType != TypeList {
		return nil, errors.New("value is not a list")
	}

	return o.listValue, nil
}

func (o *Object) ToRecord() (*Record, error) {
	if o.objectType != TypeRecord {
		return nil, errors.New("value is not a record")
	}

	return o.recordValue, nil
}

func (o *Object) ToVariable() (*Variable, error) {
	if o.objectType != TypeVariable {
		return nil, errors.New("value is not a variable")
	}

	return o.variableValue, nil
}
