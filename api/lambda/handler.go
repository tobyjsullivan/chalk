package lambda

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tobyjsullivan/chalk/api"
	"github.com/tobyjsullivan/chalk/variables"
	"google.golang.org/grpc"
	"log"
	"os"
)

func main() {
	varsSvc := os.Getenv("VARIABLES_SVC")
	conn, err := grpc.Dial(varsSvc, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial variables service: %v", err)
	}
	defer conn.Close()
	handler := api.NewHandler(variables.NewVariablesClient(conn))

	lambda.Start(handler.HandleRequest)
}
