package main

import (
	"context"
	"errors"
	"log"

	"github.com/satori/go.uuid"

	"github.com/tobyjsullivan/chalk/monolith"
)

type sessionsServer struct {
	store map[string]bool
}

func newSessionServer() *sessionsServer {
	return &sessionsServer{
		store: make(map[string]bool),
	}
}

func (s *sessionsServer) CreateSession(ctx context.Context, request *monolith.CreateSessionRequest) (*monolith.CreateSessionResponse, error) {
	log.Println("CreateSession")
	id, _ := generateSessionId()

	// TODO (toby): Track an owner?
	s.store[id] = true

	return &monolith.CreateSessionResponse{
		Session: &monolith.Session{
			SessionId: id,
		},
	}, nil
}

func (s *sessionsServer) GetSession(ctx context.Context, request *monolith.GetSessionRequest) (*monolith.GetSessionResponse, error) {
	log.Println("GetSession")
	sessId := request.Session
	if sessId == "" {
		return nil, errors.New("sessionId cannot be empty")
	}

	exists := s.store[sessId]
	if !exists {
		return &monolith.GetSessionResponse{
			Error: &monolith.Error{
				Message: "requested session does not exist " + sessId,
			},
		}, nil
	}

	return &monolith.GetSessionResponse{
		Session: &monolith.Session{
			SessionId: sessId,
		},
	}, nil
}

func generateSessionId() (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
