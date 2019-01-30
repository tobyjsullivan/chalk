package api

import (
	"context"
	"encoding/json"
	"github.com/tobyjsullivan/chalk/resolver"
	"github.com/tobyjsullivan/chalk/variables"
	"log"
	"net/http"
	"strings"
)

const allowedOrigin = "*"
const headerOrigin = "origin"

type ApiEvent struct {
	Body       string            `json:"body"`
	HttpMethod string            `json:"httpMethod"`
	Path       string            `json:"path"`
	Headers    map[string]string `json:"headers"`
}

type ApiResponse struct {
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers"`
	Body            []byte            `json:"body"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}

type Handler struct {
	varSvc variables.VariablesClient
}

func NewHandler(variablesSvc variables.VariablesClient) *Handler {
	return &Handler{
		varSvc: variablesSvc,
	}
}

func (h *Handler) HandleRequest(ctx context.Context, request *ApiEvent) (*ApiResponse, error) {
	request.Headers = normaliseHeaders(request.Headers)

	resp, err := h.varSvc.GetVariables(ctx, &variables.GetVariablesRequest{
		Keys: []string{"var1", "var2"},
	})
	if err != nil {
		log.Println("Error from GetVariables:", err)
	} else {
		log.Printf("Received: %v", resp)
	}

	switch request.HttpMethod {
	case http.MethodPost:
		return doPost(ctx, request)
	case http.MethodOptions:
		return doOptions(ctx, request)
	default:
		return &ApiResponse{
			StatusCode:      http.StatusMethodNotAllowed,
			Body:            []byte("405 Method Not Allowed"),
			IsBase64Encoded: false,
		}, nil
	}
}

func doOptions(ctx context.Context, req *ApiEvent) (*ApiResponse, error) {
	return &ApiResponse{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(req),
		Body:            []byte(""),
		IsBase64Encoded: false,
	}, nil
}

func doPost(ctx context.Context, req *ApiEvent) (*ApiResponse, error) {
	body := req.Body
	var query resolver.QueryRequest
	err := json.Unmarshal([]byte(body), &query)
	if err != nil {
		return nil, err
	}

	result := resolver.Query(&query)

	b, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &ApiResponse{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(req),
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

func determineCorsHeaders(req *ApiEvent) map[string]string {
	origin, ok := req.Headers[headerOrigin]
	if !ok || origin == "" {
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
