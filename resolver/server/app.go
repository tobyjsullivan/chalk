//go:generate protoc -I ../ --go_out=plugins=grpc:../ ../resolver.proto

package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/tobyjsullivan/chalk/variables"

	"github.com/tobyjsullivan/chalk/resolver"
	"github.com/tobyjsullivan/chalk/resolver/engine"
	"google.golang.org/grpc"
)

// server is used to implement ResolverServer.
type server struct {
	engine *engine.Engine
}

func (s *server) Resolve(ctx context.Context, in *resolver.ResolveRequest) (*resolver.ResolveResponse, error) {
	log.Println("Received:", in.Formula)
	res := s.engine.Query(ctx, in)
	log.Println("Returning:", res)
	return res, nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	varsSvc := os.Getenv("VARIABLES_SVC")

	varsConn, err := grpc.Dial(varsSvc, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial variables service: %v", err)
	}
	defer varsConn.Close()

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	resolver.RegisterResolverServer(s, &server{
		engine: engine.NewEngine(variables.NewVariablesClient(varsConn)),
	})

	log.Println("Starting server on", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
