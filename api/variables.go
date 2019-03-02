package api

type createSessionRequest struct {
}

type createSessionResponse struct {
	Session *sessionState `json:"session"`
}

type getSessionRequest struct {
}

type getSessionResponse struct {
	Error   *string       `json:"error,omitempty"`
	Session *sessionState `json:"session,omitempty"`
}

type createVariableRequest struct {
	Page    string `json:"page"`
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

type sessionState struct {
	Id    string   `json:"id"`
	Pages []string `json:"pages"`
}

type pageState struct {
	Id        string           `json:"id"`
	Variables []*variableState `json:"variables"`
}

type variableState struct {
	Id           string           `json:"id"`
	Name         string           `json:"name"`
	Formula      string           `json:"formula"`
	Result       *executionResult `json:"result"`
	Dependencies []string         `json:"dependencies"`
}
