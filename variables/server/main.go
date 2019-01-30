//go:generate protoc -I ../ --go_out=plugins=grpc:../ ../variables.proto

package main

import (
	"context"
	"errors"
	"github.com/tobyjsullivan/chalk/variables"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

// server is used to implement VariablesServer.
type server struct{}

func (s *server) GetVariables(ctx context.Context, in *variables.GetVariablesRequest) (*variables.GetVariablesResponse, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	variables.RegisterVariablesServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
