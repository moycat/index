package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/moycat/index/data"
)

type fakeRepo struct {
	reindexAllPostsFn func(ctx context.Context, req data.IndexRequest) error
	searchFn          func(ctx context.Context, query string, limit, offset int) ([]data.SearchRow, error)
}

func (f *fakeRepo) ReindexAllPosts(ctx context.Context, req data.IndexRequest) error {
	if f.reindexAllPostsFn != nil {
		return f.reindexAllPostsFn(ctx, req)
	}
	return nil
}

func (f *fakeRepo) Search(ctx context.Context, query string, limit, offset int) ([]data.SearchRow, error) {
	if f.searchFn != nil {
		return f.searchFn(ctx, query, limit, offset)
	}
	return nil, nil
}

func TestIndexServiceReindexAllPosts(t *testing.T) {
	now := time.Now().UTC()
	validRequest := data.IndexRequest{
		Posts: []data.Post{
			{
				Title:       "Hello",
				URL:         "https://example.com/p/1",
				Content:     "content",
				PublishedAt: now,
			},
		},
	}

	tests := []struct {
		name    string
		req     data.IndexRequest
		repoErr error
		wantErr bool
	}{
		{name: "ok", req: validRequest},
		{name: "validation error", req: data.IndexRequest{Posts: []data.Post{{}}}, wantErr: true},
		{name: "repository error", req: validRequest, repoErr: errors.New("db down"), wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{reindexAllPostsFn: func(ctx context.Context, req data.IndexRequest) error {
				return tc.repoErr
			}}
			svc := NewIndexService(repo)
			err := svc.ReindexAllPosts(context.Background(), tc.req)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
