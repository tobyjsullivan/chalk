package variables

import (
	"fmt"
	"sync"

	"github.com/satori/go.uuid"
)

type VariableState struct {
	Id      uuid.UUID
	Page    uuid.UUID
	Name    string
	Formula string
}

func buildVariableState(id uuid.UUID, page uuid.UUID, name string, formula string) *VariableState {
	return &VariableState{
		Id:      id,
		Page:    page,
		Name:    name,
		Formula: formula,
	}
}

type Repository interface {
	GetVariables(variableIds []uuid.UUID) ([]*VariableState, error)
	FindPageVariables(pageId uuid.UUID) []*VariableState
	FindVariablesByName(pageId uuid.UUID, names []string) []*VariableState
	CreateVariable(pageId uuid.UUID, name, formula string) (*VariableState, error)
	UpdateVariable(variableId uuid.UUID, formula string) (*VariableState, error)
	RenameVariable(variableId uuid.UUID, name string) (*VariableState, error)
}

func NewVariablesRepo() Repository {
	return &variablesRepo{
		varMap:    make(map[uuid.UUID]*VariableState),
		pageIndex: make(map[uuid.UUID][]uuid.UUID),
	}
}

type variablesRepo struct {
	mx        sync.RWMutex
	varMap    map[uuid.UUID]*VariableState
	pageIndex map[uuid.UUID][]uuid.UUID
}

func (r *variablesRepo) getVariableState(variableId uuid.UUID) *VariableState {
	r.mx.RLock()
	defer r.mx.RUnlock()
	return r.varMap[variableId]
}

func (r *variablesRepo) getPageVariableIds(pageId uuid.UUID) []uuid.UUID {
	r.mx.RLock()
	defer r.mx.RUnlock()
	variableIds := r.pageIndex[pageId]
	return variableIds
}

func (r *variablesRepo) GetVariables(variableIds []uuid.UUID) ([]*VariableState, error) {
	out := make([]*VariableState, len(variableIds))
	for i, variableId := range variableIds {
		out[i] = r.getVariableState(variableId)
		if out[i] == nil {
			return []*VariableState{}, fmt.Errorf("variable %s not found", variableId)
		}
	}

	return out, nil
}

func (r *variablesRepo) FindPageVariables(pageId uuid.UUID) []*VariableState {
	pageVars := r.getPageVariableIds(pageId)
	out := make([]*VariableState, len(pageVars))
	for i, variableId := range pageVars {
		out[i] = r.getVariableState(variableId)
	}

	return out
}

func (r *variablesRepo) FindVariablesByName(pageId uuid.UUID, names []string) []*VariableState {
	pageVars := r.getPageVariableIds(pageId)
	nameMap := make(map[string]*VariableState)

	for _, variableId := range pageVars {
		state := r.getVariableState(variableId)
		nameMap[state.Name] = state
	}

	out := make([]*VariableState, 0, len(names))
	for _, name := range names {
		if state, ok := nameMap[name]; ok {
			out = append(out, state)
		}
	}

	return out
}

func (r *variablesRepo) CreateVariable(pageId uuid.UUID, name, formula string) (*VariableState, error) {
	existing := r.FindVariablesByName(pageId, []string{name})
	if len(existing) > 0 {
		return nil, fmt.Errorf("variable `%s` already exists", name)
	}

	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	state := buildVariableState(id, pageId, name, formula)
	r.addVariable(state)

	return state, nil
}

func (r *variablesRepo) addVariable(state *VariableState) {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.varMap[state.Id] = state
	r.pageIndex[state.Page] = append(r.pageIndex[state.Page], state.Id)
}

func (r *variablesRepo) UpdateVariable(variableId uuid.UUID, formula string) (*VariableState, error) {
	state := r.getVariableState(variableId)
	if state == nil {
		return nil, fmt.Errorf("variable %s does not exist", variableId)
	}

	newState := buildVariableState(variableId, state.Page, state.Name, formula)
	r.mx.Lock()
	defer r.mx.Unlock()
	r.varMap[variableId] = newState

	return newState, nil
}
func (r *variablesRepo) RenameVariable(variableId uuid.UUID, name string) (*VariableState, error) {
	state := r.getVariableState(variableId)
	if state == nil {
		return nil, fmt.Errorf("variable %s does not exist", variableId)
	}

	newState := buildVariableState(variableId, state.Page, name, state.Formula)
	r.mx.Lock()
	defer r.mx.Unlock()
	r.varMap[variableId] = newState

	return newState, nil
}
