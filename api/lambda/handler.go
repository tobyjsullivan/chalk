package lambda

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tobyjsullivan/chalk/api"
	resolverRpc "github.com/tobyjsullivan/chalk/resolver/rpc"
	"github.com/tobyjsullivan/chalk/variables"
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

	handler := api.NewHandler(resolverRpc.NewResolverClient(resolverConn), variables.NewVariablesClient(varsConn))

	lambda.Start(handler.HandleRequest)
}
