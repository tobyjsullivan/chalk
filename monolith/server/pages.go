package main

import (
	"context"
	"log"
	"sync"

	"github.com/satori/go.uuid"

	"github.com/tobyjsullivan/chalk/monolith"
)

type pagesServer struct {
	mx         sync.RWMutex
	pages      map[uuid.UUID]*pageState
	idxSession map[uuid.UUID][]uuid.UUID
}

type pageState struct {
	session uuid.UUID
}

func newPagesServer() *pagesServer {
	return &pagesServer{
		pages:      make(map[uuid.UUID]*pageState),
		idxSession: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (s *pagesServer) CreatePage(ctx context.Context, req *monolith.CreatePageRequest) (*monolith.CreatePageResponse, error) {
	log.Println("CreatePage")
	sessId, err := uuid.FromString(req.Session)
	if err != nil {
		return nil, err
	}

	pageId, _ := uuid.NewV4()

	s.mx.Lock()
	defer s.mx.Unlock()
	s.pages[pageId] = &pageState{
		session: sessId,
	}
	s.idxSession[sessId] = append(s.idxSession[sessId], pageId)

	return &monolith.CreatePageResponse{
		Page: &monolith.Page{
			PageId:  pageId.String(),
			Session: sessId.String(),
		},
	}, nil
}

func (s *pagesServer) GetPages(ctx context.Context, req *monolith.GetPagesRequest) (*monolith.GetPagesResponse, error) {
	log.Println("GetPages")
	out := make([]*monolith.Page, 0, len(req.PageIds))

	s.mx.RLock()
	defer s.mx.RUnlock()
	for _, p := range req.PageIds {
		pageId, err := uuid.FromString(p)
		if err != nil {
			return nil, err
		}

		state, ok := s.pages[pageId]
		if !ok {
			continue
		}

		out = append(out, &monolith.Page{
			PageId:  pageId.String(),
			Session: state.session.String(),
		})
	}

	return &monolith.GetPagesResponse{
		Pages: out,
	}, nil
}

func (s *pagesServer) FindPages(ctx context.Context, req *monolith.FindPagesRequest) (*monolith.FindPagesResponse, error) {
	log.Println("FindPages")
	sessId, err := uuid.FromString(req.Session)
	if err != nil {
		return nil, err
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
			PageId:  pageId.String(),
			Session: sessId.String(),
		}
	}

	return &monolith.FindPagesResponse{
		Pages: pages,
	}, nil
}
