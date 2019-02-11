package main

import (
	"context"
	"log"

	"github.com/tobyjsullivan/chalk/monolith"
)

// variablesServer is used to implement VariablesServer.
type variablesServer struct {
	varMap map[string]string
}

func (s *variablesServer) GetVariables(ctx context.Context, in *monolith.GetVariablesRequest) (*monolith.GetVariablesResponse, error) {
	var out []*monolith.Variable
	for _, k := range in.Keys {
		f := s.varMap[k]
		out = append(out, &monolith.Variable{
			Name:    k,
			Formula: f,
		})
		log.Println("Sending var", k, ":", f)
	}

	return &monolith.GetVariablesResponse{
		Values: out,
	}, nil
}

func (s *variablesServer) SetVariable(ctx context.Context, in *monolith.SetVariableRequest) (*monolith.SetVariableResponse, error) {
	key := in.Key
	value := in.Formula
	log.Println("Setting var", key, ":", value)

	if value == "" {
		delete(s.varMap, key)
	} else {
		s.varMap[key] = value
	}

	return &monolith.SetVariableResponse{
		Variable: &monolith.Variable{},
	}, nil
}
