package main

import (
	"context"

	"github.com/tobyjsullivan/chalk/monolith"
)

type sessionsServer struct {
}

func (s *sessionsServer) GetSession(ctx context.Context, request *monolith.GetSessionRequest) (*monolith.GetSessionResponse, error) {
	// TODO
	return &monolith.GetSessionResponse{
		Session: &monolith.Session{
			SessionId: "",
			Pages:     []string{},
		},
	}, nil
}
