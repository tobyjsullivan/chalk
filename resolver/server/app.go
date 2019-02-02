//go:generate protoc -I ../ --go_out=plugins=grpc:../ ../resolver.proto

package main

import (
	"context"
	"github.com/tobyjsullivan/chalk/resolver"
	"github.com/tobyjsullivan/chalk/resolver/engine"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

// server is used to implement ResolverServer.
type server struct{}

func (s *server) Resolve(ctx context.Context, in *resolver.ResolveRequest) (*resolver.ResolveResponse, error) {
	log.Println("Received:", in.Formula)
	res := engine.Query(in)
	log.Println("Returning:", res)
	return res, nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	resolver.RegisterResolverServer(s, &server{})

	log.Println("Starting server on", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
