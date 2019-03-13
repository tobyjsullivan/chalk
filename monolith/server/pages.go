package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"sync"

	"github.com/tobyjsullivan/chalk/monolith"
)

// Should be a multiple of 3 to avoid padding with `=` during base64 encoding.
const pageIdSizeBytes = 9

type pagesServer struct {
	mx         sync.RWMutex
	pages      map[string]*pageState
	idxSession map[string][]string
}

type pageState struct {
	session string
}

func newPagesServer() *pagesServer {
	return &pagesServer{
		pages:      make(map[string]*pageState),
		idxSession: make(map[string][]string),
	}
}

func (s *pagesServer) CreatePage(ctx context.Context, req *monolith.CreatePageRequest) (*monolith.CreatePageResponse, error) {
	log.Println("CreatePage")

	sessId := req.Session
	if sessId == "" {
		return nil, errors.New("sessionId cannot be empty")
	}
	pageId, _ := generatePageId()

	s.mx.Lock()
	defer s.mx.Unlock()
	s.pages[pageId] = &pageState{
		session: sessId,
	}
	s.idxSession[sessId] = append(s.idxSession[sessId], pageId)

	return &monolith.CreatePageResponse{
		Page: &monolith.Page{
			PageId:  pageId,
			Session: sessId,
		},
	}, nil
}

func (s *pagesServer) GetPages(ctx context.Context, req *monolith.GetPagesRequest) (*monolith.GetPagesResponse, error) {
	log.Println("GetPages")
	out := make([]*monolith.Page, 0, len(req.PageIds))

	s.mx.RLock()
	defer s.mx.RUnlock()
	for _, pageId := range req.PageIds {
		if pageId == "" {
			return nil, errors.New("pageId cannot be empty")
		}

		state, ok := s.pages[pageId]
		if !ok {
			continue
		}

		out = append(out, &monolith.Page{
			PageId:  pageId,
			Session: state.session,
		})
	}

	return &monolith.GetPagesResponse{
		Pages: out,
	}, nil
}

func (s *pagesServer) FindPages(ctx context.Context, req *monolith.FindPagesRequest) (*monolith.FindPagesResponse, error) {
	log.Println("FindPages")
	sessId := req.Session
	if sessId == "" {
		return nil, errors.New("sessionId cannot be empty")
	}

	s.mx.RLock()
	defer s.mx.RUnlock()

	pageIds, ok := s.idxSession[sessId]
	if !ok {
		log.Println("no pages for session:", sessId)
		return &monolith.FindPagesResponse{
			Pages: []*monolith.Page{},
		}, nil
	}

	pages := make([]*monolith.Page, len(pageIds))
	for i, pageId := range pageIds {
		pages[i] = &monolith.Page{
			PageId:  pageId,
			Session: sessId,
		}
	}

	return &monolith.FindPagesResponse{
		Pages: pages,
	}, nil
}

func generatePageId() (string, error) {
	pid := [pageIdSizeBytes]byte{}
	if _, err := rand.Read(pid[:]); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(pid[:]), nil
}
