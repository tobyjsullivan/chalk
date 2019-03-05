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
	resolverSvcHost := os.Getenv("RESOLVER_SVC")
	monolithHost := os.Getenv("VARIABLES_SVC")

	resolverConn, err := grpc.Dial(resolverSvcHost, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial resolver service: %v", err)
	}
	defer resolverConn.Close()

	monolithConn, err := grpc.Dial(monolithHost, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial variables service: %v", err)
	}
	defer monolithConn.Close()

	pagesSvc := monolith.NewPagesClient(monolithConn)
	resolverSvc := resolverRpc.NewResolverClient(resolverConn)
	sessionsSvc := monolith.NewSessionsClient(monolithConn)
	variablesSvc := monolith.NewVariablesClient(monolithConn)
	handler := api.NewHandler(pagesSvc, resolverSvc, sessionsSvc, variablesSvc)

	lambda.Start(handler.HandleRequest)
}
