//go:generate protoc -I ../ --go_out=plugins=grpc:../ ../domain.proto
//go:generate protoc -I ../ --go_out=plugins=grpc:../ ../pages.proto
//go:generate protoc -I ../ --go_out=plugins=grpc:../ ../sessions.proto
//go:generate protoc -I ../ --go_out=plugins=grpc:../ ../variables.proto

package main

import (
	"log"
	"net"
	"os"

	"github.com/tobyjsullivan/chalk/monolith"
	"google.golang.org/grpc"
)

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

	log.Println("Registering VariablesServer...")
	monolith.RegisterVariablesServer(s, newVariablesServer())
	log.Println("Registering SessionsServer...")
	monolith.RegisterSessionsServer(s, newSessionServer())
	log.Println("Registering PagesServer...")
	monolith.RegisterPagesServer(s, newPagesServer())

	log.Println("Starting server on", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
