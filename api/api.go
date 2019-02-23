package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/tobyjsullivan/chalk/monolith"
	"github.com/tobyjsullivan/chalk/resolver"
)

const allowedOrigin = "*"
const headerOrigin = "origin"

var (
	reVariablesCollection = regexp.MustCompile("^/variables$")
	reVariablesDocument   = regexp.MustCompile("^/variables/([a-fA-F0-9-]+)$")

	rePathCreateVar = reVariablesCollection
	rePathUpdateVar = reVariablesDocument
	rePathGetVars   = reVariablesCollection
)

type Event struct {
	Body                            string              `json:"body"`
	HttpMethod                      string              `json:"httpMethod"`
	MultiValueQueryStringParameters map[string][]string `json:"multiValueQueryStringParameters"`
	Path                            string              `json:"path"`
	Headers                         map[string]string   `json:"headers"`
}

type Response struct {
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers"`
	Body            []byte            `json:"body"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}

type Handler struct {
	resolverSvc  resolver.ResolverClient
	variablesSvc monolith.VariablesClient
}

func NewHandler(resolverSvc resolver.ResolverClient, variablesSvc monolith.VariablesClient) *Handler {
	return &Handler{
		resolverSvc:  resolverSvc,
		variablesSvc: variablesSvc,
	}
}

func (h *Handler) HandleRequest(ctx context.Context, request *Event) (*Response, error) {
	switch request.HttpMethod {
	case http.MethodGet:
		return h.doGet(ctx, request)
	case http.MethodPost:
		return h.doPost(ctx, request)
	case http.MethodOptions:
		return h.doOptions(ctx, request)
	default:
		return &Response{
			StatusCode:      http.StatusMethodNotAllowed,
			Body:            []byte("405 Method Not Allowed"),
			IsBase64Encoded: false,
		}, nil
	}
}

func (h *Handler) doOptions(ctx context.Context, req *Event) (*Response, error) {
	return &Response{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(req),
		Body:            []byte(""),
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) doGet(ctx context.Context, req *Event) (*Response, error) {
	if req.Path == "/health" {
		return h.doGetHealth(ctx, req)
	} else if rePathGetVars.MatchString(req.Path) {
		return h.doGetVariables(ctx, req)
	}

	return &Response{
		StatusCode:      http.StatusNotFound,
		Body:            []byte("404 Not Found"),
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) doPost(ctx context.Context, req *Event) (*Response, error) {
	if rePathCreateVar.MatchString(req.Path) {
		return h.doCreateVariable(ctx, req)
	} else if rePathUpdateVar.MatchString(req.Path) {
		return h.doUpdateVariable(ctx, req)
	}

	return &Response{
		StatusCode:      http.StatusNotFound,
		Body:            []byte("404 Not Found"),
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) doGetHealth(ctx context.Context, req *Event) (*Response, error) {
	return &Response{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(req),
		Body:            []byte("{}"),
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) doCreateVariable(ctx context.Context, event *Event) (*Response, error) {
	body := event.Body
	var createRequest createVariableRequest
	err := json.Unmarshal([]byte(body), &createRequest)
	if err != nil {
		return nil, err
	}

	resp, err := h.variablesSvc.SetVariable(ctx, &monolith.SetVariableRequest{
		Name:    createRequest.Name,
		Formula: createRequest.Formula,
	})
	if err != nil {
		return nil, err
	}

	var out createVariableResponse
	if resp.Error != nil {
		out.Error = &resp.Error.Message
	} else if resp.Variable != nil {
		out.State, err = h.buildVariableState(ctx, resp.Variable)
		if err != nil {
			return nil, err
		}
	}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(event),
		Body:            b,
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) doUpdateVariable(ctx context.Context, event *Event) (*Response, error) {
	body := event.Body
	var updateRequest updateVariableRequest
	err := json.Unmarshal([]byte(body), &updateRequest)
	if err != nil {
		return nil, err
	}

	matches := rePathUpdateVar.FindStringSubmatch(event.Path)
	if len(matches) != 2 {
		panic("Expected ID in path.")
	}

	id := matches[1]
	varReq := &monolith.SetVariableRequest{
		Id: id,
	}
	if updateRequest.Name != nil {
		varReq.Name = *updateRequest.Name
	}
	if updateRequest.Formula != nil {
		varReq.Formula = *updateRequest.Formula
	}

	resp, err := h.variablesSvc.SetVariable(ctx, varReq)
	if err != nil {
		return nil, err
	}

	var out updateVariableResponse
	if resp.Error != nil {
		out.Error = &resp.Error.Message
	}

	if resp.Variable != nil {
		out.State, err = h.buildVariableState(ctx, resp.Variable)
		if err != nil {
			return nil, err
		}
	}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(event),
		Body:            b,
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) buildVariableState(ctx context.Context, v *monolith.Variable) (*variableState, error) {
	state := &variableState{
		Id:      v.VariableId,
		Name:    v.Name,
		Formula: v.Formula,
	}

	// Execute the formula
	formula := v.Formula
	result, err := h.resolverSvc.Resolve(ctx, &resolver.ResolveRequest{
		Formula: formula,
	})
	if err != nil {
		return nil, err
	}

	state.Result, err = mapResolveResponse(result)

	return state, nil
}

func (h *Handler) doGetVariables(ctx context.Context, event *Event) (*Response, error) {
	ids := event.MultiValueQueryStringParameters["id"]
	resp, err := h.variablesSvc.GetVariables(ctx, &monolith.GetVariablesRequest{
		Ids: ids,
	})
	if err != nil {
		return nil, err
	}

	var out getVariablesResponse
	out.Variables = make([]*variableState, len(resp.Values))
	for i, v := range resp.Values {
		out.Variables[i], err = h.buildVariableState(ctx, v)
		if err != nil {
			return nil, err
		}
	}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(event),
		Body:            b,
		IsBase64Encoded: false,
	}, nil
}

func normaliseHeaders(in map[string]string) map[string]string {
	out := make(map[string]string)

	for k, v := range in {
		norm := strings.ToLower(k)
		out[norm] = v
	}

	return out
}

func determineCorsHeaders(req *Event) map[string]string {
	origin, ok := normaliseHeaders(req.Headers)[headerOrigin]
	if !ok || origin == "" {
		log.Println("No origin")
		return map[string]string{}
	}

	if allowedOrigin != "*" && strings.ToLower(allowedOrigin) != strings.ToLower(origin) {
		return map[string]string{}
	}

	headers := make(map[string]string)

	switch req.HttpMethod {
	case http.MethodOptions:
		headers["Access-Control-Allow-Methods"] = "POST, OPTIONS"
		headers["Access-Control-Allow-Headers"] = "content-type"
		headers["Access-Control-Max-Age"] = "86400"
		fallthrough
	default:
		headers["Access-Control-Allow-origin"] = origin
	}

	return headers
}
