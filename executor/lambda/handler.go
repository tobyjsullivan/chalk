package lambda

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tobyjsullivan/chalk/api"
	"github.com/tobyjsullivan/chalk/variables"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"
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
	Body            []byte	          `json:"body"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}

type Handler struct {
	varSvc variables.VariablesClient
}

func (h *Handler) HandleRequest(ctx context.Context, request *ApiEvent) (*ApiResponse, error) {
	request.Headers = normaliseHeaders(request.Headers)

	switch request.HttpMethod {
	case http.MethodPost:
		return doPost(ctx, request)
	case http.MethodOptions:
		return doOptions(ctx, request)
	default:
		return &ApiResponse{
			StatusCode: http.StatusMethodNotAllowed,
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
	var query api.QueryRequest
	err := json.Unmarshal([]byte(body), &query)
	if err != nil {
		return nil, err
	}

	result := api.Query(&query)

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

func main() {
	varsSvc := os.Getenv("VARIABLES_SVC")
	conn, err := grpc.Dial(varsSvc, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial variables service: %v", err)
	}
	defer conn.Close()
	handler := Handler{
		varSvc: variables.NewVariablesClient(conn),
	}

	lambda.Start(handler.HandleRequest)
}
