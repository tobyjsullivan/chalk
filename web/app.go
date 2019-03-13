package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/tobyjsullivan/chalk/monolith"
	"google.golang.org/grpc"
)

const cookieCurrentSession = "CURRENT_SESSION"

type bootstrapObject struct {
	PageId string `json:"page_id"`
}

type pageTemplateVariables struct {
	Bootstrap *bootstrapObject
}

type handler struct {
	pagesSvc    monolith.PagesClient
	sessionsSvc monolith.SessionsClient
}

func newHandler(pagesSvc monolith.PagesClient, sessionsSvc monolith.SessionsClient) http.Handler {
	return &handler{
		pagesSvc:    pagesSvc,
		sessionsSvc: sessionsSvc,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request:", r.Method, r.URL.Path)

	// Serve static files under /static
	if strings.HasPrefix(r.URL.Path, "/static") {
		fs := http.FileServer(http.Dir("/webapp"))
		http.StripPrefix("/static", fs).ServeHTTP(w, r)
		return
	}

	if r.URL.Path == "/" {
		h.handleRoot(w, r)
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

func (h *handler) handleRoot(w http.ResponseWriter, r *http.Request) {
	// Redirect requests for root to a newly created page
	ctx := r.Context()
	// Check cookies for session or else create a new one.
	var sessionId string
	if c, err := r.Cookie(cookieCurrentSession); err == http.ErrNoCookie {
		// Create a new session
		sessResp, err := h.sessionsSvc.CreateSession(ctx, &monolith.CreateSessionRequest{})
		if err != nil {
			log.Println("error in FindPages:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		sessionId = sessResp.Session.SessionId
		http.SetCookie(w, &http.Cookie{
			Name:  cookieCurrentSession,
			Value: sessionId,
		})
	} else {
		sessionId = c.Value
	}

	var pageId string
	pagesResp, err := h.pagesSvc.FindPages(ctx, &monolith.FindPagesRequest{
		Session: sessionId,
	})
	if err != nil {
		log.Println("error in FindPages:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if pagesResp.Error != nil {
		log.Println("error from FindPages:", pagesResp.Error.Message)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(pagesResp.Pages) == 0 {
		// Create a new page
		createPageResp, err := h.pagesSvc.CreatePage(ctx, &monolith.CreatePageRequest{
			Session: sessionId,
		})
		if err != nil {
			log.Println("error in FindPages:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if createPageResp.Error != nil {
			log.Println("error from FindPages:", pagesResp.Error.Message)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		pageId = createPageResp.Page.PageId
	} else {
		// Default page
		pageId = pagesResp.Pages[0].PageId
	}

	http.Redirect(w, r, "/"+pageId, http.StatusTemporaryRedirect)
	return
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sessionsSvcHost := os.Getenv("SESSIONS_SVC")
	pagesSvcHost := os.Getenv("PAGES_SVC")

	sessionsConn, err := grpc.Dial(sessionsSvcHost, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial variables service: %v", err)
	}
	defer sessionsConn.Close()
	log.Println("dialed sessions svc:", sessionsSvcHost)
	sessionsSvc := monolith.NewSessionsClient(sessionsConn)

	pagesConn, err := grpc.Dial(pagesSvcHost, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial variables service: %v", err)
	}
	defer pagesConn.Close()
	log.Println("dialed pages svc:", pagesSvcHost)
	pagesSvc := monolith.NewPagesClient(sessionsConn)

	server := http.Server{
		Addr:    ":" + port,
		Handler: newHandler(pagesSvc, sessionsSvc),
	}

	log.Println("Starting on port", port)
	log.Fatal(server.ListenAndServe())
}
