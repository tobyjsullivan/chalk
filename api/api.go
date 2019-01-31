package api

import (
	"context"
	"encoding/json"
	resolver_rpc "github.com/tobyjsullivan/chalk/resolver/rpc"
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
	resolverSvc  resolver_rpc.ResolverClient
	variablesSvc variables.VariablesClient
}

func NewHandler(resolverSvc resolver_rpc.ResolverClient, variablesSvc variables.VariablesClient) *Handler {
	return &Handler{
		resolverSvc:  resolverSvc,
		variablesSvc: variablesSvc,
	}
}

func (h *Handler) HandleRequest(ctx context.Context, request *ApiEvent) (*ApiResponse, error) {
	resp, err := h.variablesSvc.GetVariables(ctx, &variables.GetVariablesRequest{
		Keys: []string{"var1", "var2"},
	})
	if err != nil {
		log.Println("Error from GetVariables:", err)
	} else {
		log.Printf("Received: %v", resp)
	}

	switch request.HttpMethod {
	case http.MethodPost:
		return h.doPost(ctx, request)
	case http.MethodOptions:
		return h.doOptions(ctx, request)
	default:
		return &ApiResponse{
			StatusCode:      http.StatusMethodNotAllowed,
			Body:            []byte("405 Method Not Allowed"),
			IsBase64Encoded: false,
		}, nil
	}
}

func (h *Handler) doOptions(ctx context.Context, req *ApiEvent) (*ApiResponse, error) {
	return &ApiResponse{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(req),
		Body:            []byte(""),
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) doPost(ctx context.Context, req *ApiEvent) (*ApiResponse, error) {
	body := req.Body
	var query resolver_rpc.ResolveRequest
	err := json.Unmarshal([]byte(body), &query)
	if err != nil {
		return nil, err
	}

	result, err := h.resolverSvc.Resolve(ctx, &query)
	if err != nil {
		return nil, err
	}

	var out executionResult
	out.Error = result.Error
	if result.Result != nil {
		out.Result = &executionResultObject{}
		switch result.Result.Type {
		case resolver_rpc.ObjectType_STRING:
			out.Result.Type = "string"
			out.Result.StringValue = result.Result.StringValue
		case resolver_rpc.ObjectType_NUMBER:
			out.Result.Type = "number"
			out.Result.NumberValue = result.Result.NumberValue
		}
	}

	b, err := json.Marshal(out)
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

type executionResult struct {
	Result *executionResultObject `json:"result,omitempty'"`
	Error  string                 `json:"error,omitempty"`
}

type executionResultObject struct {
	Type        string  `json:"type"`
	NumberValue float64 `json:"numberValue,omitempty"`
	StringValue string  `json:"stringValue,omitempty"`
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
