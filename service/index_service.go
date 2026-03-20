package service

import (
	"context"
	"fmt"

	"github.com/moycat/index/data"
)

type IndexService struct {
	repo PostRepository
}

func NewIndexService(repo PostRepository) *IndexService {
	return &IndexService{repo: repo}
}

func (s *IndexService) ReindexAllPosts(ctx context.Context, req data.IndexRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidArgument, err)
	}
	if err := s.repo.ReindexAllPosts(ctx, req); err != nil {
		return fmt.Errorf("reindex all posts: %w", err)
	}
	return nil
}
