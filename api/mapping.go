package api

import (
	"fmt"

	"github.com/tobyjsullivan/chalk/resolver"
)

type executionResult struct {
	Result *executionResultObject `json:"result,omitempty'"`
	Error  string                 `json:"error,omitempty"`
}

type executionResultObject struct {
	Type         *executionResultObjectType `json:"type"`
	BooleanValue *bool                      `json:"booleanValue,omitempty"`
	LambdaValue  *executionResultLambda     `json:"lambdaValue,omitempty"`
	ListValue    *executionResultList       `json:"listValue,omitempty"`
	NumberValue  *float64                   `json:"numberValue,omitempty"`
	RecordValue  *executionResultRecord     `json:"recordValue,omitempty"`
	StringValue  *string                    `json:"stringValue,omitempty"`
}

type executionResultObjectType struct {
	Class string `json:"class"`
}

type executionResultLambda struct {
	FreeVariables []string `json:"freeVariables"`
}

type executionResultList struct {
	Elements []*executionResultObject `json:"elements"`
}

type executionResultRecord struct {
	Properties map[string]*executionResultObject `json:"properties"`
}

func mapResolveResponse(resp *resolver.ResolveResponse) (*executionResult, error) {
	out := &executionResult{}
	out.Error = resp.Error
	if resp.Result != nil {
		var err error
		out.Result, err = mapResolveResponseObject(resp.Result)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func mapResolveResponseObject(object *resolver.Object) (*executionResultObject, error) {
	switch object.Type {
	case resolver.ObjectType_BOOLEAN:
		return &executionResultObject{
			Type: &executionResultObjectType{
				Class: "boolean",
			},
			BooleanValue: &object.BoolValue,
		}, nil
	case resolver.ObjectType_LAMBDA:
		freeVars := object.LambdaValue.FreeVariables

		result := &executionResultObject{
			Type: &executionResultObjectType{
				Class: "lambda",
			},
			LambdaValue: &executionResultLambda{
				FreeVariables: make([]string, len(freeVars)),
			},
		}

		copy(result.LambdaValue.FreeVariables, freeVars)

		return result, nil
	case resolver.ObjectType_LIST:
		elements := object.ListValue.Elements
		listObj := &executionResultObject{
			Type: &executionResultObjectType{
				Class: "list",
			},
			ListValue: &executionResultList{
				Elements: make([]*executionResultObject, len(elements)),
			},
		}
		var err error
		for i, e := range elements {
			listObj.ListValue.Elements[i], err = mapResolveResponseObject(e)
			if err != nil {
				return nil, err
			}
		}

		return listObj, nil
	case resolver.ObjectType_NUMBER:
		return &executionResultObject{
			Type: &executionResultObjectType{
				Class: "number",
			},
			NumberValue: &object.NumberValue,
		}, nil
	case resolver.ObjectType_RECORD:
		recordObj := &executionResultObject{
			Type: &executionResultObjectType{
				Class: "record",
			},
			RecordValue: &executionResultRecord{
				Properties: make(map[string]*executionResultObject),
			},
		}

		var err error
		for _, prop := range object.RecordValue.Properties {
			recordObj.RecordValue.Properties[prop.Name], err = mapResolveResponseObject(prop.Value)
			if err != nil {
				return nil, err
			}
		}

		return recordObj, nil
	case resolver.ObjectType_STRING:
		return &executionResultObject{
			Type: &executionResultObjectType{
				Class: "string",
			},
			StringValue: &object.StringValue,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected result type: %s", object.Type)
	}
}
