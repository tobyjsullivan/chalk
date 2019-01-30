package main

import (
	"context"
	"github.com/tobyjsullivan/chalk/executor/lambda"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	headers := make(map[string]string)
	for k, vs := range r.Header {
		if len(vs) > 0 {
			headers[k] = vs[0]
		}
	}

	req := &lambda.ApiEvent{
		Body:       string(body),
		HttpMethod: r.Method,
		Path:       r.URL.Path,
		Headers:    headers,
	}

	handler := lambda.Handler{}
	ctx := context.Background()
	resp, err := handler.HandleRequest(ctx, req)
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
	s := &http.Server{
		Addr:           ":" + port,
		Handler:        &handler{},
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Println("Starting on", port)
	log.Fatal(s.ListenAndServe())
}
