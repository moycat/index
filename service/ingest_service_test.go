package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/moycat/index/data"
)

type fakeRepo struct {
	replaceSnapshotFn func(ctx context.Context, snapshot data.Snapshot) error
	searchFn          func(ctx context.Context, query string, limit, offset int) ([]data.SearchRow, error)
}

func (f *fakeRepo) ReplaceSnapshot(ctx context.Context, snapshot data.Snapshot) error {
	if f.replaceSnapshotFn != nil {
		return f.replaceSnapshotFn(ctx, snapshot)
	}
	return nil
}

func (f *fakeRepo) Search(ctx context.Context, query string, limit, offset int) ([]data.SearchRow, error) {
	if f.searchFn != nil {
		return f.searchFn(ctx, query, limit, offset)
	}
	return nil, nil
}

func TestIngestServiceReplaceSnapshot(t *testing.T) {
	now := time.Now().UTC()
	validSnapshot := data.Snapshot{
		SnapshotID:  "snapshot-1",
		GeneratedAt: now,
		Posts: []data.Post{
			{
				ID:          "post-1",
				Title:       "Hello",
				URL:         "https://example.com/p/1",
				Content:     "content",
				PublishedAt: now,
			},
		},
	}

	tests := []struct {
		name     string
		snapshot data.Snapshot
		repoErr  error
		wantErr  bool
	}{
		{name: "ok", snapshot: validSnapshot},
		{name: "validation error", snapshot: data.Snapshot{}, wantErr: true},
		{name: "repository error", snapshot: validSnapshot, repoErr: errors.New("db down"), wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{replaceSnapshotFn: func(ctx context.Context, snapshot data.Snapshot) error {
				return tc.repoErr
			}}
			svc := NewIngestService(repo)
			err := svc.ReplaceSnapshot(context.Background(), tc.snapshot)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
