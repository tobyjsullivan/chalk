package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tobyjsullivan/chalk/api"
)

func handleRequest(ctx context.Context, request *api.QueryRequest) (*api.QueryResult, error) {
	return api.Query(request), nil
}

func main() {
	lambda.Start(handleRequest)
}
