package lambda

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tobyjsullivan/chalk/api"
	"github.com/tobyjsullivan/chalk/monolith"
	resolverRpc "github.com/tobyjsullivan/chalk/resolver"
	"google.golang.org/grpc"
)

func main() {
	resolverSvc := os.Getenv("RESOLVER_SVC")
	varsSvc := os.Getenv("VARIABLES_SVC")

	resolverConn, err := grpc.Dial(resolverSvc, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial resolver service: %v", err)
	}
	defer resolverConn.Close()

	varsConn, err := grpc.Dial(varsSvc, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial variables service: %v", err)
	}
	defer varsConn.Close()

	handler := api.NewHandler(resolverRpc.NewResolverClient(resolverConn), monolith.NewVariablesClient(varsConn))

	lambda.Start(handler.HandleRequest)
}
