//go:generate protoc -I ../ --go_out=plugins=grpc:../ ../variables.proto

package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/tobyjsullivan/chalk/variables"
	"google.golang.org/grpc"
)

// server is used to implement VariablesServer.
type server struct {
	varMap map[string]string
}

func (s *server) GetVariables(ctx context.Context, in *variables.GetVariablesRequest) (*variables.GetVariablesResponse, error) {
	log.Printf("Received GetVariables request: %v", in)
	var out []*variables.GetVariablesResponse_VariableEntry
	for _, k := range in.Keys {
		f := s.varMap[k]
		out = append(out, &variables.GetVariablesResponse_VariableEntry{
			Key:     k,
			Formula: f,
		})
	}

	return &variables.GetVariablesResponse{
		Values: out,
	}, nil
}

func (s *server) SetVariable(ctx context.Context, in *variables.SetVariableRequest) (*variables.VoidResponse, error) {
	key := in.Key
	value := in.Formula
	if value == "" {
		delete(s.varMap, key)
	} else {
		s.varMap[key] = value
	}

	return &variables.VoidResponse{}, nil
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
	variables.RegisterVariablesServer(s, &server{
		varMap: make(map[string]string),
	})

	log.Println("Starting server on", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
