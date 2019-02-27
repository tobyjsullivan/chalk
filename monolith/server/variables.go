package main

import (
	"context"
	"fmt"
	"log"
	"strings"
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
		id := s.findVariableId(name)
		if id != nil {
			ids = append(ids, *id)
		}
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
	var id uuid.UUID
	var err error
	if in.Id == "" {
		id, err = s.createVariable(in.Name, in.Formula)
		if err != nil {
			return nil, err
		}
	} else {
		id, err = uuid.FromString(in.Id)
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
	}

	state, err := s.getVariable(id)
	if err != nil {
		return nil, err
	}

	return &monolith.SetVariableResponse{
		Variable: &monolith.Variable{
			VariableId: id.String(),
			Name:       state.name,
			Formula:    state.formula,
		},
	}, nil
}

func (s *variablesServer) createVariable(name string, formula string) (uuid.UUID, error) {
	log.Println("createVariable:", name, formula)
	err := validateName(name)
	if err != nil {
		return uuid.UUID{}, err
	}

	// Check for existing.
	existingId := s.findVariableId(name)
	if existingId != nil {
		return *existingId, s.updateVariable(*existingId, formula)
	}

	id, err := uuid.NewV4()
	if err != nil {
		return uuid.UUID{}, err
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	name = normalizeVarName(name)
	s.varMap[id] = &state{
		name:    name,
		formula: formula,
	}
	s.nameIndex[name] = id

	return id, nil
}

func (s *variablesServer) renameVariable(id uuid.UUID, name string) error {
	log.Println("renameVariable:", id, name)
	err := validateName(name)
	if err != nil {
		return err
	}

	existingId := s.findVariableId(name)
	if existingId != nil {
		return fmt.Errorf("variable `%s` already exists", name)
	}

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
	s.nameIndex[normalizeVarName(name)] = id

	return nil
}

func (s *variablesServer) updateVariable(id uuid.UUID, formula string) error {
	log.Println("updateVariable:", id, formula)
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
	log.Println("getVariable:", id)
	s.mx.RLock()
	defer s.mx.RUnlock()

	state, ok := s.varMap[id]
	if !ok {
		return nil, fmt.Errorf("variable not found: %s", id)
	}

	return state, nil
}

func (s *variablesServer) findVariableId(name string) *uuid.UUID {
	s.mx.RLock()
	defer s.mx.RUnlock()

	id, ok := s.nameIndex[normalizeVarName(name)]
	if !ok {
		// It is reasonable that a search by name will not match. This is not an error. Return nil.
		return nil
	}

	return &id
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
