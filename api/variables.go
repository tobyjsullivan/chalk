package api

type createVariableRequest struct {
	Name    string `json:"name"`
	Formula string `json:"formula"`
}

type createVariableResponse struct {
	Error *string        `json:"error,omitempty"`
	State *variableState `json:"state,omitempty"`
}

type updateVariableRequest struct {
	Name    *string `json:"name"`
	Formula *string `json:"formula"`
}

type updateVariableResponse struct {
	Error *string        `json:"error,omitempty"`
	State *variableState `json:"state,omitempty"`
}

type getVariablesResponse struct {
	Variables []*variableState `json:"variables"`
}

type variableState struct {
	Id           string           `json:"id"`
	Name         string           `json:"name"`
	Formula      string           `json:"formula"`
	Result       *executionResult `json:"result"`
	Dependencies []string         `json:"dependencies"`
}
