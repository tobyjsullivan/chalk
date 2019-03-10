package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

type bootstrapObject struct {
	PageId string `json:"page_id"`
}

type pageTemplateVariables struct {
	Bootstrap *bootstrapObject
}

type handler struct {
}

func (*handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/static") {
		fs := http.FileServer(http.Dir("/webapp"))
		http.StripPrefix("/static", fs).ServeHTTP(w, r)
		return
	}

	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Println("error parsing template:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, &pageTemplateVariables{
		Bootstrap: &bootstrapObject{
			PageId: r.URL.Path[len("/"):],
		},
	})
	if err != nil {
		log.Println("error executing template:", err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := http.Server{
		Addr:    ":" + port,
		Handler: &handler{},
	}

	log.Println("Starting on port", port)
	log.Fatal(server.ListenAndServe())
}
