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
	varMap    map[uuid.UUID]*varState
	pageIndex map[uuid.UUID][]uuid.UUID
}

func newVariablesServer() *variablesServer {
	return &variablesServer{
		varMap:    make(map[uuid.UUID]*varState),
		pageIndex: make(map[uuid.UUID][]uuid.UUID),
	}
}

type varState struct {
	id      uuid.UUID
	page    uuid.UUID
	name    string
	formula string
}

func (s *variablesServer) GetVariables(ctx context.Context, in *monolith.GetVariablesRequest) (*monolith.GetVariablesResponse, error) {
	log.Println("GetVariables")
	out := make([]*monolith.Variable, len(in.Ids))
	for i, vid := range in.Ids {
		id, err := uuid.FromString(vid)
		if err != nil {
			return nil, err
		}

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

func (s *variablesServer) FindVariables(ctx context.Context, in *monolith.FindVariablesRequest) (*monolith.FindVariablesResponse, error) {
	log.Println("FindVariables")
	pageId, err := uuid.FromString(in.PageId)
	if err != nil {
		return nil, err
	}

	var out []*monolith.Variable
	if len(in.Names) > 0 {
		out, err = s.findVarsByName(pageId, in.Names)
		if err != nil {
			return nil, err
		}
	} else {
		varStates := s.findPageVariables(pageId)
		out := make([]*monolith.Variable, len(varStates))
		for i, s := range varStates {
			out[i] = &monolith.Variable{
				VariableId: s.id.String(),
				Name:       s.name,
				Formula:    s.formula,
			}
		}
	}

	return &monolith.FindVariablesResponse{
		Values: out,
	}, nil
}

func (s *variablesServer) CreateVariable(ctx context.Context, in *monolith.CreateVariableRequest) (*monolith.CreateVariableResponse, error) {
	log.Println("CreateVariable")
	pageId, err := uuid.FromString(in.PageId)
	if err != nil {
		return nil, err
	}

	id, err := s.createVariable(pageId, in.Name, in.Formula)
	if err != nil {
		return nil, err
	}

	state, err := s.getVariable(id)
	if err != nil {
		return nil, err
	}

	return &monolith.CreateVariableResponse{
		Variable: &monolith.Variable{
			VariableId: id.String(),
			Page:       pageId.String(),
			Name:       state.name,
			Formula:    state.formula,
		},
	}, nil
}

func (s *variablesServer) UpdateVariable(ctx context.Context, in *monolith.UpdateVariableRequest) (*monolith.UpdateVariableResponse, error) {
	log.Println("UpdateVariable")
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

	state, err := s.getVariable(id)
	if err != nil {
		return nil, err
	}

	return &monolith.UpdateVariableResponse{
		Variable: &monolith.Variable{
			VariableId: id.String(),
			Page:       state.page.String(),
			Name:       state.name,
			Formula:    state.formula,
		},
	}, nil
}

func (s *variablesServer) createVariable(page uuid.UUID, name string, formula string) (uuid.UUID, error) {
	log.Println("createVariable:", name, formula)
	err := validateName(name)
	if err != nil {
		return uuid.UUID{}, err
	}

	// Check for existing.
	if varState := s.findVariableByName(page, name); varState != nil {
		return varState.id, s.updateVariable(varState.id, formula)
	}

	id, err := uuid.NewV4()
	if err != nil {
		return uuid.UUID{}, err
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	name = normalizeVarName(name)
	s.varMap[id] = &varState{
		page:    page,
		name:    name,
		formula: formula,
	}
	s.pageIndex[page] = append(s.pageIndex[page], id)

	return id, nil
}

func (s *variablesServer) renameVariable(id uuid.UUID, name string) error {
	log.Println("renameVariable:", id, name)
	err := validateName(name)
	if err != nil {
		return err
	}

	s.mx.RLock()
	state := s.varMap[id]
	s.mx.RUnlock()

	if varState := s.findVariableByName(state.page, name); varState != nil {
		if varState.id == id {
			// no-op
			return nil
		}
		return fmt.Errorf("variable `%s` already exists", name)
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	oldState, ok := s.varMap[id]
	if !ok {
		return fmt.Errorf("variable not found: %s", id)
	}

	s.varMap[id] = &varState{
		name:    name,
		formula: oldState.formula,
	}

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

	s.varMap[id] = &varState{
		name:    oldState.name,
		formula: formula,
	}

	return nil
}

func (s *variablesServer) getVariable(id uuid.UUID) (*varState, error) {
	log.Println("getVariable:", id)
	s.mx.RLock()
	defer s.mx.RUnlock()

	state, ok := s.varMap[id]
	if !ok {
		return nil, fmt.Errorf("variable not found: %s", id)
	}

	return state, nil
}

func (s *variablesServer) findVarsByName(pageId uuid.UUID, names []string) ([]*monolith.Variable, error) {
	out := make([]*monolith.Variable, 0, len(names))
	for _, name := range names {
		varState := s.findVariableByName(pageId, name)
		if varState == nil {
			continue
		}

		out = append(out, &monolith.Variable{
			VariableId: varState.id.String(),
			Name:       varState.name,
			Formula:    varState.formula,
		})

		log.Println("Sending var", varState.name, ":", varState.formula)
	}

	return out, nil
}

func (s *variablesServer) findPageVariables(page uuid.UUID) []*varState {
	log.Println("findPageVariables:", page)

	s.mx.RLock()
	defer s.mx.RUnlock()

	pageVars, ok := s.pageIndex[page]
	if !ok {
		return []*varState{}
	}

	out := make([]*varState, len(pageVars))
	for i, v := range pageVars {
		out[i] = s.varMap[v]
	}

	return out
}

func (s *variablesServer) findVariableByName(page uuid.UUID, name string) *varState {
	log.Println("findVariableId:", page, name)

	pageVarStates := s.findPageVariables(page)

	for _, state := range pageVarStates {
		if state.name == name {
			return state
		}
	}

	return nil
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
