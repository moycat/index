package service

import (
	"context"
	"fmt"

	"github.com/moycat/index/data"
)

type IngestService struct {
	repo PostRepository
}

func NewIngestService(repo PostRepository) *IngestService {
	return &IngestService{repo: repo}
}

func (s *IngestService) ReplaceSnapshot(ctx context.Context, snapshot data.Snapshot) error {
	if err := snapshot.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidArgument, err)
	}
	if err := s.repo.ReplaceSnapshot(ctx, snapshot); err != nil {
		return fmt.Errorf("replace snapshot: %w", err)
	}
	return nil
}
