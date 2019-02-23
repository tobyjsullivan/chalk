package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/satori/go.uuid"

	"github.com/tobyjsullivan/chalk/monolith"
)

// variablesServer is used to implement VariablesServer.
type variablesServer struct {
	mx        sync.RWMutex
	varMap    map[uuid.UUID]*state
	nameIndex map[string]uuid.UUID
}

func newVariablesServer() *variablesServer {
	return &variablesServer{
		varMap:    make(map[uuid.UUID]*state),
		nameIndex: make(map[string]uuid.UUID),
	}
}

type state struct {
	name    string
	formula string
}

func (s *variablesServer) GetVariables(ctx context.Context, in *monolith.GetVariablesRequest) (*monolith.GetVariablesResponse, error) {
	ids := make([]uuid.UUID, len(in.Ids))
	var err error
	for i, id := range in.Ids {
		ids[i], err = uuid.FromString(id)
		if err != nil {
			return nil, err
		}
	}

	for _, name := range in.Names {
		id, err := s.findVariableId(name)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	out := make([]*monolith.Variable, len(ids))
	for i, id := range ids {
		state, err := s.getVariable(id)
		if err != nil {
			return nil, err
		}

		out[i] = &monolith.Variable{
			VariableId: id.String(),
			Name:       state.name,
			Formula:    state.formula,
		}

		log.Println("Sending var", state.name, ":", state.formula)
	}

	return &monolith.GetVariablesResponse{
		Values: out,
	}, nil
}

func (s *variablesServer) SetVariable(ctx context.Context, in *monolith.SetVariableRequest) (*monolith.SetVariableResponse, error) {
	id, err := uuid.FromString(in.Id)
	if err != nil {
		return nil, err
	}

	if in.Name != "" {
		err := s.renameVariable(id, in.Name)
		if err != nil {
			return nil, err
		}
	}

	if in.Formula != "" {
		err := s.updateVariable(id, in.Formula)
		if err != nil {
			return nil, err
		}
	}

	return &monolith.SetVariableResponse{
		Variable: &monolith.Variable{
			VariableId: id.String(),
			Name:       in.Name,
			Formula:    in.Formula,
		},
	}, nil
}

func (s *variablesServer) createVariable(name string, formula string) (uuid.UUID, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return uuid.UUID{}, err
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	s.varMap[id] = &state{
		name:    name,
		formula: formula,
	}
	s.nameIndex[name] = id

	return id, nil
}

func (s *variablesServer) renameVariable(id uuid.UUID, name string) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	oldState, ok := s.varMap[id]
	if !ok {
		return fmt.Errorf("variable not found: %s", id)
	}

	s.varMap[id] = &state{
		name:    name,
		formula: oldState.formula,
	}
	delete(s.nameIndex, oldState.name)
	s.nameIndex[name] = id

	return nil
}

func (s *variablesServer) updateVariable(id uuid.UUID, formula string) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	oldState, ok := s.varMap[id]
	if !ok {
		return fmt.Errorf("variable not found: %s", id)
	}

	s.varMap[id] = &state{
		name:    oldState.name,
		formula: formula,
	}

	return nil
}

func (s *variablesServer) getVariable(id uuid.UUID) (*state, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	state, ok := s.varMap[id]
	if !ok {
		return nil, fmt.Errorf("variable not found: %s", id)
	}

	return state, nil
}

func (s *variablesServer) findVariableId(name string) (uuid.UUID, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	id, ok := s.nameIndex[name]
	if !ok {
		return uuid.UUID{}, fmt.Errorf("variable not found: %s", id)
	}

	return id, nil
}
