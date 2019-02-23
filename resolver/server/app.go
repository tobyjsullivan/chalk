//go:generate protoc -I ../ --go_out=plugins=grpc:../ ../resolver.proto

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/tobyjsullivan/chalk/resolver/engine/types"

	"github.com/tobyjsullivan/chalk/monolith"

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
	obj, err := s.engine.Query(ctx, in.Formula)
	var res *resolver.ResolveResponse
	if err != nil {
		res = toErrorResult(err)
	} else {
		res = toResult(obj)
	}
	log.Println("Returning:", res)
	return res, nil
}

func toResult(res *types.Object) *resolver.ResolveResponse {
	obj, err := toResultObject(res)
	if err != nil {
		return toErrorResult(err)
	}

	return &resolver.ResolveResponse{
		Result: obj,
	}
}

func toResultObject(obj *types.Object) (*resolver.Object, error) {
	if obj == nil {
		return nil, nil
	}
	switch obj.Type() {
	case types.TypeNumber:
		n, _ := obj.ToNumber()
		return &resolver.Object{
			Type:        resolver.ObjectType_NUMBER,
			NumberValue: n,
		}, nil
	case types.TypeString:
		s, _ := obj.ToString()
		return &resolver.Object{
			Type:        resolver.ObjectType_STRING,
			StringValue: s,
		}, nil
	case types.TypeList:
		list, _ := obj.ToList()

		els := make([]*resolver.Object, len(list.Elements))
		var err error
		for i, el := range list.Elements {
			els[i], err = toResultObject(el)
			if err != nil {
				return nil, err
			}
		}

		return &resolver.Object{
			Type: resolver.ObjectType_LIST,
			ListValue: &resolver.List{
				Elements: els,
			},
		}, nil
	case types.TypeRecord:
		record, _ := obj.ToRecord()

		props := make([]*resolver.RecordProperty, 0, len(record.Properties))
		for k, v := range record.Properties {
			value, err := toResultObject(v)
			if err != nil {
				return nil, err
			}

			props = append(props, &resolver.RecordProperty{
				Name:  k,
				Value: value,
			})
		}

		return &resolver.Object{
			Type: resolver.ObjectType_RECORD,
			RecordValue: &resolver.Record{
				Properties: props,
			},
		}, nil
	case types.TypeLambda:
		lambda, _ := obj.ToLambda()

		return &resolver.Object{
			Type: resolver.ObjectType_LAMBDA,
			LambdaValue: &resolver.Lambda{
				FreeVariables: lambda.FreeVariables,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected result type: %v", obj.Type())
	}
}

func toErrorResult(err error) *resolver.ResolveResponse {
	return &resolver.ResolveResponse{
		Error: fmt.Sprint(err),
	}
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
		engine: engine.NewEngine(monolith.NewVariablesClient(varsConn)),
	})

	log.Println("Starting server on", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
