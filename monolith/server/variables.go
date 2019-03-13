package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tobyjsullivan/chalk/monolith/server/variables"

	"github.com/tobyjsullivan/chalk/monolith"
)

const maxBatchSize = 100

// variablesServer is used to implement VariablesServer.
type variablesServer struct {
	repo variables.Repository
}

func newVariablesServer() *variablesServer {
	return &variablesServer{
		repo: variables.NewVariablesRepo(),
	}
}

func (s *variablesServer) GetVariables(ctx context.Context, in *monolith.GetVariablesRequest) (*monolith.GetVariablesResponse, error) {
	log.Println("GetVariables")

	states, err := s.repo.GetVariables(in.Ids)
	if err != nil {
		return nil, err
	}

	out := make([]*monolith.Variable, len(states))
	for i, state := range states {
		out[i] = &monolith.Variable{
			VariableId: state.Id,
			Page:       state.Page,
			Name:       state.Name,
			Formula:    state.Formula,
		}

		log.Println("Sending var", state.Name, ":", state.Formula)
	}

	return &monolith.GetVariablesResponse{
		Values: out,
	}, nil
}

func (s *variablesServer) FindVariables(ctx context.Context, in *monolith.FindVariablesRequest) (*monolith.FindVariablesResponse, error) {
	log.Println("FindVariables")
	pageId := in.PageId
	if pageId == "" {
		return nil, errors.New("pageId cannot be empty")
	}

	var states []*variables.VariableState
	if l := len(in.Names); l > 0 {
		if l > maxBatchSize {
			return nil, fmt.Errorf("max batch size exceeded: %d", l)
		}

		// normalise names
		names := make([]string, len(in.Names))
		for i, n := range in.Names {
			names[i] = normalizeVarName(n)
		}

		states = s.repo.FindVariablesByName(pageId, names)
	} else {
		states = s.repo.FindPageVariables(pageId)
	}
	log.Println("found", len(states), "variables")
	out := make([]*monolith.Variable, len(states))
	for i, s := range states {
		out[i] = &monolith.Variable{
			VariableId: s.Id,
			Page:       s.Page,
			Name:       s.Name,
			Formula:    s.Formula,
		}
	}

	return &monolith.FindVariablesResponse{
		Values: out,
	}, nil
}

func (s *variablesServer) CreateVariable(ctx context.Context, in *monolith.CreateVariableRequest) (*monolith.CreateVariableResponse, error) {
	log.Println("CreateVariable")
	pageId := in.PageId
	if pageId == "" {
		return nil, errors.New("pageId cannot be empty")
	}

	name := normalizeVarName(in.Name)
	err := validateName(name)
	if err != nil {
		return nil, err
	}

	state, err := s.repo.CreateVariable(pageId, name, in.Formula)
	if err != nil {
		return nil, err
	}

	return &monolith.CreateVariableResponse{
		Variable: &monolith.Variable{
			VariableId: state.Id,
			Page:       state.Page,
			Name:       state.Name,
			Formula:    state.Formula,
		},
	}, nil
}

func (s *variablesServer) UpdateVariable(ctx context.Context, in *monolith.UpdateVariableRequest) (*monolith.UpdateVariableResponse, error) {
	log.Println("UpdateVariable")
	id := in.Id
	if id == "" {
		return nil, errors.New("pageId cannot be empty")
	}

	var err error
	var state *variables.VariableState
	if in.Name != "" {
		name := normalizeVarName(in.Name)
		err = validateName(name)
		if err != nil {
			return nil, err
		}

		state, err = s.repo.RenameVariable(id, name)
		if err != nil {
			return nil, err
		}
	}

	if in.Formula != "" {
		state, err = s.repo.UpdateVariable(id, in.Formula)
		if err != nil {
			return nil, err
		}
	}

	return &monolith.UpdateVariableResponse{
		Variable: &monolith.Variable{
			VariableId: state.Id,
			Page:       state.Page,
			Name:       state.Name,
			Formula:    state.Formula,
		},
	}, nil
}

func normalizeVarName(name string) string {
	return strings.ToLower(name)
}

func validateName(name string) error {
	if strings.ContainsAny(name, " -_") {
		return fmt.Errorf("variable name `%s` contains invalid characters", name)
	}

	if name[0] >= '0' && name[0] <= '9' {
		return fmt.Errorf("variable name `%s` must not start with a number", name)
	}

	return nil
}
