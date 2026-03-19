package service

import (
	"context"
	"testing"
	"time"

	"github.com/moycat/index/adapter/index/ngram"
	"github.com/moycat/index/adapter/index/snippet"
	"github.com/moycat/index/data"
)

func TestSearchServiceSearch(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{
		searchFn: func(ctx context.Context, query string, limit, offset int) ([]data.SearchRow, error) {
			return []data.SearchRow{
				{
					Post: data.Post{
						ID:          "1",
						Title:       "Go 和中文检索",
						URL:         "https://example.com/1",
						Content:     "这是一篇关于 Golang 全文检索的文章。",
						PublishedAt: now,
					},
					Score: 1.5,
				},
			}, nil
		},
	}

	svc := NewSearchService(repo, ngram.NewTokenizer(2), snippet.NewBuilder(), 20)
	hits, err := svc.Search(context.Background(), "中文 检索", 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].Title == "" || hits[0].URL == "" || hits[0].Snippet == "" {
		t.Fatalf("expected non-empty hit fields")
	}
	if len(hits[0].MatchedTerms) == 0 {
		t.Fatalf("expected matched terms")
	}
}

func TestSearchServiceSearchInvalidQuery(t *testing.T) {
	svc := NewSearchService(&fakeRepo{}, ngram.NewTokenizer(2), snippet.NewBuilder(), 20)
	_, err := svc.Search(context.Background(), "   ", 1, 10)
	if err == nil {
		t.Fatalf("expected error for empty query")
	}
}
