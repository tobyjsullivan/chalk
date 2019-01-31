//go:generate protoc -I ../rpc/ --go_out=plugins=grpc:../rpc/ ../rpc/resolver.proto

package main

import (
	"context"
	"github.com/tobyjsullivan/chalk/resolver"
	"github.com/tobyjsullivan/chalk/resolver/rpc"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

// server is used to implement ResolverServer.
type server struct{}

func (s *server) Resolve(ctx context.Context, in *rpc.ResolveRequest) (*rpc.ResolveResponse, error) {
	log.Println("Received:", in.Formula)
	res := resolver.Query(in)
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
	rpc.RegisterResolverServer(s, &server{})

	log.Println("Starting server on", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
