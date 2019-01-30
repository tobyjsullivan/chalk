package main

import (
	"context"
	"github.com/tobyjsullivan/chalk/executor"
	"github.com/tobyjsullivan/chalk/variables"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type handler struct {
	executionHandler *executor.Handler
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	headers := make(map[string]string)
	for k, vs := range r.Header {
		if len(vs) > 0 {
			headers[k] = vs[0]
		}
	}

	req := &executor.ApiEvent{
		Body:       string(body),
		HttpMethod: r.Method,
		Path:       r.URL.Path,
		Headers:    headers,
	}

	ctx := context.Background()
	resp, err := h.executionHandler.HandleRequest(ctx, req)
	if err != nil {
		log.Panicf("Error handling request: %v", err)
	}

	w.WriteHeader(resp.StatusCode)
	for k, v := range resp.Headers {
		w.Header().Add(k, v)
	}
	w.Write(resp.Body)
}

func main() {
	port := "8080"

	varsSvc := os.Getenv("VARIABLES_SVC")
	conn, err := grpc.Dial(varsSvc, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial variables service: %v", err)
	}
	defer conn.Close()

	s := &http.Server{
		Addr: ":" + port,
		Handler: &handler{
			executionHandler: executor.NewHandler(variables.NewVariablesClient(conn)),
		},
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Println("Starting on", port)
	log.Fatal(s.ListenAndServe())
}
