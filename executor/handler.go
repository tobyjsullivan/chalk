package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tobyjsullivan/chalk/api"
	"net/http"
)

type ApiEvent struct {
	Body       string            `json:"body"`
	HttpMethod string            `json:"httpMethod"`
	Path       string            `json:"path"`
	Headers    map[string]string `json:"headers"`
}

type ApiResponse struct {
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers"`
	Body            interface{}       `json:"body"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}

func handleRequest(ctx context.Context, request *ApiEvent) (*ApiResponse, error) {
	switch request.HttpMethod {
	case http.MethodPost:
		return doPost(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported method %s", request.HttpMethod)
	}
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
		Body:            string(b),
		IsBase64Encoded: false,
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
