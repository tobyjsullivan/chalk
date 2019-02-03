package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tobyjsullivan/chalk/api"
	"github.com/tobyjsullivan/chalk/resolver"
	"github.com/tobyjsullivan/chalk/monolith"
	"google.golang.org/grpc"
)

type handler struct {
	executionHandler *api.Handler
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request", r.Method, r.URL.Path)
	body, _ := ioutil.ReadAll(r.Body)
	headers := make(map[string]string)
	for k, vs := range r.Header {
		if len(vs) > 0 {
			headers[k] = vs[0]
		}
	}

	req := &api.ApiEvent{
		Body:       string(body),
		HttpMethod: r.Method,
		Path:       r.URL.Path,
		Headers:    headers,
	}

	ctx := context.Background()
	resp, err := h.executionHandler.HandleRequest(ctx, req)
	if err != nil {
		log.Printf("Error handling request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	for k, v := range resp.Headers {
		w.Header().Add(k, v)
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(resp.Body)
}

func main() {
	port := "8080"

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

	s := &http.Server{
		Addr: ":" + port,
		Handler: &handler{
			executionHandler: api.NewHandler(
				resolver.NewResolverClient(resolverConn),
				monolith.NewVariablesClient(varsConn),
			),
		},
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Println("Starting on", port)
	log.Fatal(s.ListenAndServe())
}
