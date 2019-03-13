package variables

import (
	"fmt"
	"sync"

	"github.com/satori/go.uuid"
)

type VariableState struct {
	Id      string
	Page    string
	Name    string
	Formula string
}

func buildVariableState(id string, page string, name string, formula string) *VariableState {
	return &VariableState{
		Id:      id,
		Page:    page,
		Name:    name,
		Formula: formula,
	}
}

type Repository interface {
	GetVariables(variableIds []string) ([]*VariableState, error)
	FindPageVariables(pageId string) []*VariableState
	FindVariablesByName(pageId string, names []string) []*VariableState
	CreateVariable(pageId, name, formula string) (*VariableState, error)
	UpdateVariable(variableId, formula string) (*VariableState, error)
	RenameVariable(variableId, name string) (*VariableState, error)
}

func NewVariablesRepo() Repository {
	return &variablesRepo{
		varMap:    make(map[string]*VariableState),
		pageIndex: make(map[string][]string),
	}
}

type variablesRepo struct {
	mx        sync.RWMutex
	varMap    map[string]*VariableState
	pageIndex map[string][]string
}

func (r *variablesRepo) getVariableState(variableId string) *VariableState {
	r.mx.RLock()
	defer r.mx.RUnlock()
	return r.varMap[variableId]
}

func (r *variablesRepo) getPageVariableIds(pageId string) []string {
	r.mx.RLock()
	defer r.mx.RUnlock()
	variableIds := r.pageIndex[pageId]
	return variableIds
}

func (r *variablesRepo) GetVariables(variableIds []string) ([]*VariableState, error) {
	out := make([]*VariableState, len(variableIds))
	for i, variableId := range variableIds {
		out[i] = r.getVariableState(variableId)
		if out[i] == nil {
			return []*VariableState{}, fmt.Errorf("variable %s not found", variableId)
		}
	}

	return out, nil
}

func (r *variablesRepo) FindPageVariables(pageId string) []*VariableState {
	pageVars := r.getPageVariableIds(pageId)
	out := make([]*VariableState, len(pageVars))
	for i, variableId := range pageVars {
		out[i] = r.getVariableState(variableId)
	}

	return out
}

func (r *variablesRepo) FindVariablesByName(pageId string, names []string) []*VariableState {
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

func (r *variablesRepo) CreateVariable(pageId, name, formula string) (*VariableState, error) {
	existing := r.FindVariablesByName(pageId, []string{name})
	if len(existing) > 0 {
		return nil, fmt.Errorf("variable `%s` already exists", name)
	}

	id, err := generateVariableId()
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

func (r *variablesRepo) UpdateVariable(variableId, formula string) (*VariableState, error) {
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
func (r *variablesRepo) RenameVariable(variableId, name string) (*VariableState, error) {
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

func generateVariableId() (string, error) {
	if id, err := uuid.NewV4(); err != nil {
		return "", err
	} else {
		return id.String(), nil
	}
}
