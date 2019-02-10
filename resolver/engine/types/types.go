package types

import "errors"

const (
	TypeApplication TypeName = "application"
	TypeList                 = "list"
	TypeNumber               = "number"
	TypeRecord               = "record"
	TypeString               = "string"
	TypeVariable             = "variable"
)

type TypeName string

type Application struct {
	FunctionName string
	Arguments    []*Object
}

type List struct {
	Elements []*Object
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
	listValue        *List
	numberValue      float64
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

func NewApplication(funcName string, args []*Object) *Object {
	return &Object{
		objectType: TypeApplication,
		applicationValue: &Application{
			FunctionName: funcName,
			Arguments:    args,
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
	if o.objectType != TypeString {
		return "", errors.New("value is not a string")
	}

	return o.stringValue, nil
}

func (o *Object) ToNumber() (float64, error) {
	if o.objectType != TypeNumber {
		return 0, errors.New("value is not a number")
	}

	return o.numberValue, nil
}

func (o *Object) ToApplication() (*Application, error) {
	if o.objectType != TypeApplication {
		return nil, errors.New("value is not an application")
	}

	return o.applicationValue, nil
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
