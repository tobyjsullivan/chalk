package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"sync"

	"github.com/satori/go.uuid"

	"github.com/tobyjsullivan/chalk/monolith"
)

// Should be a multiple of 3 to avoid padding with `=` during base64 encoding.
const pageIdSizeBytes = 9

type pageId [pageIdSizeBytes]byte

type pagesServer struct {
	mx         sync.RWMutex
	pages      map[pageId]*pageState
	idxSession map[uuid.UUID][]pageId
}

type pageState struct {
	session uuid.UUID
}

func newPagesServer() *pagesServer {
	return &pagesServer{
		pages:      make(map[pageId]*pageState),
		idxSession: make(map[uuid.UUID][]pageId),
	}
}

func (s *pagesServer) CreatePage(ctx context.Context, req *monolith.CreatePageRequest) (*monolith.CreatePageResponse, error) {
	log.Println("CreatePage")
	sessId, err := uuid.FromString(req.Session)
	if err != nil {
		return nil, err
	}

	pageId, _ := generatePid()

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
		pageId, err := pidFromString(p)
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

func generatePid() (pageId, error) {
	pid := pageId{}
	if _, err := rand.Read(pid[:]); err != nil {
		return pageId{}, err
	}

	return pid, nil
}

func pidFromString(s string) (pageId, error) {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return pageId{}, err
	}

	// For now we want exact sizes. It seems reasonable that longer sizes will be needed in future.
	// The primary risk right now would be the unintended use of shorter strings (eg, if someone visits /about).
	if n := len(b); n != pageIdSizeBytes {
		return pageId{}, fmt.Errorf("unexpected pageId lenght: %d bytes", n)
	}

	pid := pageId{}
	copy(pid[:], b)

	return pid, nil
}

func (pid pageId) String() string {
	return base64.URLEncoding.EncodeToString(pid[:])
}
