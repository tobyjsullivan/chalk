package main

import (
	"context"
	"log"

	"github.com/satori/go.uuid"

	"github.com/tobyjsullivan/chalk/monolith"
)

type sessionsServer struct {
	store map[uuid.UUID]bool
}

func newSessionServer() *sessionsServer {
	return &sessionsServer{
		store: make(map[uuid.UUID]bool),
	}
}

func (s *sessionsServer) CreateSession(ctx context.Context, request *monolith.CreateSessionRequest) (*monolith.CreateSessionResponse, error) {
	log.Println("CreateSession")
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	// TODO (toby): Track an owner?
	s.store[id] = true

	return &monolith.CreateSessionResponse{
		Session: &monolith.Session{
			SessionId: id.String(),
		},
	}, nil
}

func (s *sessionsServer) GetSession(ctx context.Context, request *monolith.GetSessionRequest) (*monolith.GetSessionResponse, error) {
	log.Println("GetSession")
	sessId, err := uuid.FromString(request.Session)
	if err != nil {
		return nil, err
	}

	exists := s.store[sessId]
	if !exists {
		return &monolith.GetSessionResponse{
			Error: &monolith.Error{
				Message: "requested session does not exist " + sessId.String(),
			},
		}, nil
	}

	return &monolith.GetSessionResponse{
		Session: &monolith.Session{
			SessionId: sessId.String(),
		},
	}, nil
}
