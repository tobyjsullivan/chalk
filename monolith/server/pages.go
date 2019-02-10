package main

import (
	"context"

	"github.com/tobyjsullivan/chalk/monolith"
)

type pagesServer struct {
}

func (*pagesServer) CreatePage(context.Context, *monolith.CreatePageRequest) (*monolith.CreatePageResponse, error) {
	// TODO
	return &monolith.CreatePageResponse{
		Error: &monolith.Error{
			Message: "not implemented",
		},
	}, nil
}

func (*pagesServer) GetPages(context.Context, *monolith.GetPagesRequest) (*monolith.GetPagesResponse, error) {
	// TODO
	return &monolith.GetPagesResponse{
		Error: &monolith.Error{
			Message: "not implemented",
		},
	}, nil
}
