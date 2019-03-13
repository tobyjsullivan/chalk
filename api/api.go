package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/tobyjsullivan/chalk/monolith"
	"github.com/tobyjsullivan/chalk/resolver"
)

const allowedOrigin = "*"
const headerOrigin = "origin"

var (
	reSessionsCollection  = regexp.MustCompile("^/sessions$")
	reSessionsDocument    = regexp.MustCompile("^/sessions/([a-zA-Z0-9-_]+)$")
	reVariablesCollection = regexp.MustCompile("^/variables$")
	reVariablesDocument   = regexp.MustCompile("^/variables/([a-zA-Z0-9-_]+)$")

	rePathCreateSession    = reSessionsCollection
	rePathGetSession       = reSessionsDocument
	rePathGetPageVariables = regexp.MustCompile("^/pages/([a-zA-Z0-9-_]+)/variables$")
	rePathCreateVar        = reVariablesCollection
	rePathUpdateVar        = reVariablesDocument
)

type Event struct {
	Body                            string              `json:"body"`
	HttpMethod                      string              `json:"httpMethod"`
	MultiValueQueryStringParameters map[string][]string `json:"multiValueQueryStringParameters"`
	Path                            string              `json:"path"`
	Headers                         map[string]string   `json:"headers"`
}

type Response struct {
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers"`
	Body            []byte            `json:"body"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}

type Handler struct {
	pagesSvc     monolith.PagesClient
	resolverSvc  resolver.ResolverClient
	sessionsSvc  monolith.SessionsClient
	variablesSvc monolith.VariablesClient
}

func NewHandler(
	pagesSvc monolith.PagesClient,
	resolverSvc resolver.ResolverClient,
	sessionSvc monolith.SessionsClient,
	variablesSvc monolith.VariablesClient,
) *Handler {
	return &Handler{
		pagesSvc:     pagesSvc,
		resolverSvc:  resolverSvc,
		sessionsSvc:  sessionSvc,
		variablesSvc: variablesSvc,
	}
}

func (h *Handler) HandleRequest(ctx context.Context, request *Event) (*Response, error) {
	switch request.HttpMethod {
	case http.MethodGet:
		return h.doGet(ctx, request)
	case http.MethodPost:
		return h.doPost(ctx, request)
	case http.MethodOptions:
		return h.doOptions(ctx, request)
	default:
		return &Response{
			StatusCode:      http.StatusMethodNotAllowed,
			Body:            []byte("405 Method Not Allowed"),
			IsBase64Encoded: false,
		}, nil
	}
}

func (h *Handler) doOptions(ctx context.Context, req *Event) (*Response, error) {
	return &Response{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(req),
		Body:            []byte(""),
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) doGet(ctx context.Context, req *Event) (*Response, error) {
	if req.Path == "/health" {
		return h.doGetHealth(ctx, req)
	} else if rePathGetSession.MatchString(req.Path) {
		return h.doGetSession(ctx, req)
	} else if rePathGetPageVariables.MatchString(req.Path) {
		return h.doGetPageVariables(ctx, req)
	}

	return &Response{
		StatusCode:      http.StatusNotFound,
		Body:            []byte("404 Not Found"),
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) doPost(ctx context.Context, req *Event) (*Response, error) {
	if rePathCreateSession.MatchString(req.Path) {
		log.Println("Creating session")
		return h.doCreateSession(ctx, req)
	} else if rePathCreateVar.MatchString(req.Path) {
		log.Println("Creating var")
		return h.doCreateVariable(ctx, req)
	} else if rePathUpdateVar.MatchString(req.Path) {
		log.Println("Updating var")
		return h.doUpdateVariable(ctx, req)
	}

	return &Response{
		StatusCode:      http.StatusNotFound,
		Body:            []byte("404 Not Found"),
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) doGetHealth(ctx context.Context, req *Event) (*Response, error) {
	return &Response{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(req),
		Body:            []byte("{}"),
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) doCreateVariable(ctx context.Context, event *Event) (*Response, error) {
	body := event.Body
	var createRequest createVariableRequest
	err := json.Unmarshal([]byte(body), &createRequest)
	if err != nil {
		return nil, err
	}

	if createRequest.Page == "" {
		return &Response{
			StatusCode:      http.StatusBadRequest,
			Headers:         determineCorsHeaders(event),
			Body:            []byte("must specify page id"),
			IsBase64Encoded: false,
		}, nil
	}

	if createRequest.Name == "" {
		return &Response{
			StatusCode:      http.StatusBadRequest,
			Headers:         determineCorsHeaders(event),
			Body:            []byte("must specify variable name"),
			IsBase64Encoded: false,
		}, nil
	}

	resp, err := h.variablesSvc.CreateVariable(ctx, &monolith.CreateVariableRequest{
		PageId:  createRequest.Page,
		Name:    createRequest.Name,
		Formula: createRequest.Formula,
	})
	if err != nil {
		return nil, err
	}

	var out createVariableResponse
	if resp.Error != nil {
		out.Error = &resp.Error.Message
	} else if resp.Variable != nil {
		out.State, err = h.buildVariableState(ctx, resp.Variable)
		if err != nil {
			return nil, err
		}
	}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(event),
		Body:            b,
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) doUpdateVariable(ctx context.Context, event *Event) (*Response, error) {
	body := event.Body
	var updateRequest updateVariableRequest
	err := json.Unmarshal([]byte(body), &updateRequest)
	if err != nil {
		return nil, err
	}

	matches := rePathUpdateVar.FindStringSubmatch(event.Path)
	if len(matches) != 2 {
		// Panic because the router should have verified this previously.
		panic("Expected ID in path.")
	}

	id := matches[1]
	log.Println("Updating", id, ";", updateRequest.Name, ";", updateRequest.Formula)
	varReq := &monolith.UpdateVariableRequest{
		Id: id,
	}
	if updateRequest.Name != nil {
		varReq.Name = *updateRequest.Name
	}
	if updateRequest.Formula != nil {
		varReq.Formula = *updateRequest.Formula
	}

	resp, err := h.variablesSvc.UpdateVariable(ctx, varReq)
	if err != nil {
		return nil, err
	}

	var out updateVariableResponse
	if resp.Error != nil {
		out.Error = &resp.Error.Message
	}

	if resp.Variable != nil {
		out.State, err = h.buildVariableState(ctx, resp.Variable)
		if err != nil {
			return nil, err
		}
	}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(event),
		Body:            b,
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) buildVariableState(ctx context.Context, v *monolith.Variable) (*variableState, error) {
	state := &variableState{
		Id:      v.VariableId,
		Name:    v.Name,
		Formula: v.Formula,
	}

	// Execute the formula
	formula := v.Formula
	result, err := h.resolverSvc.Resolve(ctx, &resolver.ResolveRequest{
		PageId:  v.Page,
		Formula: formula,
	})
	if err != nil {
		return nil, err
	}

	state.Result, err = mapResolveResponse(result)

	return state, nil
}

func (h *Handler) doCreateSession(ctx context.Context, event *Event) (*Response, error) {
	sessResp, err := h.sessionsSvc.CreateSession(ctx, &monolith.CreateSessionRequest{})
	if err != nil {
		return nil, err
	}

	sessId := sessResp.Session.SessionId
	var out createSessionResponse
	out.Session = &sessionState{
		Id: sessId,
	}

	// Create default page
	pageResp, err := h.pagesSvc.CreatePage(ctx, &monolith.CreatePageRequest{
		Session: sessId,
	})
	if err != nil {
		return nil, err
	}

	defaultPageId := pageResp.Page.PageId
	out.Session.Pages = []string{defaultPageId}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(event),
		Body:            b,
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) doGetSession(ctx context.Context, event *Event) (*Response, error) {
	matches := rePathGetSession.FindStringSubmatch(event.Path)
	if len(matches) != 2 {
		// Panic because the router should have verified this previously.
		panic("Expected ID in path.")
	}

	// Session
	id := matches[1]
	sessReq := &monolith.GetSessionRequest{
		Session: id,
	}
	resp, err := h.sessionsSvc.GetSession(ctx, sessReq)
	if err != nil {
		return nil, err
	}
	var out getSessionResponse
	if resp.Error != nil {
		out.Error = &resp.Error.Message
	}
	if resp.Session != nil {
		// Get Pages
		pagesRes, err := h.pagesSvc.FindPages(ctx, &monolith.FindPagesRequest{
			Session: resp.Session.SessionId,
		})
		if err != nil {
			return nil, err
		}
		if pagesRes.Error != nil {
			out.Error = &resp.Error.Message
		}

		pages := make([]string, len(pagesRes.Pages))
		for i, p := range pagesRes.Pages {
			pages[i] = p.PageId
		}

		out.Session = &sessionState{
			Id:    resp.Session.SessionId,
			Pages: pages,
		}
	}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(event),
		Body:            b,
		IsBase64Encoded: false,
	}, nil
}

func (h *Handler) doGetPageVariables(ctx context.Context, event *Event) (*Response, error) {
	matches := rePathGetPageVariables.FindStringSubmatch(event.Path)
	if len(matches) != 2 {
		// Panic because the router should have verified this previously.
		panic("Expected ID in path.")
	}

	// Page ID
	pageId := matches[1]
	varsReq := &monolith.FindVariablesRequest{
		PageId: pageId,
	}
	resp, err := h.variablesSvc.FindVariables(ctx, varsReq)
	if err != nil {
		return nil, err
	}

	var out getPageVariablesResponse
	out.Variables = make([]*variableState, len(resp.Values))
	log.Println("found", len(resp.Values), "variables")
	for i, v := range resp.Values {
		log.Println("adding var to resp", v.VariableId)
		out.Variables[i], err = h.buildVariableState(ctx, v)
		if err != nil {
			return nil, err
		}
	}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode:      http.StatusOK,
		Headers:         determineCorsHeaders(event),
		Body:            b,
		IsBase64Encoded: false,
	}, nil
}

func normaliseHeaders(in map[string]string) map[string]string {
	out := make(map[string]string)

	for k, v := range in {
		norm := strings.ToLower(k)
		out[norm] = v
	}

	return out
}

func determineCorsHeaders(req *Event) map[string]string {
	origin, ok := normaliseHeaders(req.Headers)[headerOrigin]
	if !ok || origin == "" {
		log.Println("No origin")
		return map[string]string{}
	}

	if allowedOrigin != "*" && strings.ToLower(allowedOrigin) != strings.ToLower(origin) {
		return map[string]string{}
	}

	headers := make(map[string]string)

	switch req.HttpMethod {
	case http.MethodOptions:
		headers["Access-Control-Allow-Methods"] = "POST, OPTIONS"
		headers["Access-Control-Allow-Headers"] = "content-type"
		headers["Access-Control-Max-Age"] = "86400"
		fallthrough
	default:
		headers["Access-Control-Allow-origin"] = origin
	}

	return headers
}
