package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/moycat/index/adapter/index/ngram"
	"github.com/moycat/index/adapter/index/snippet"
	"github.com/moycat/index/data"
	"github.com/moycat/index/service"
	log "github.com/sirupsen/logrus"
)

type testRepo struct {
	posts map[string]data.Post
}

func newTestRepo() *testRepo {
	return &testRepo{posts: make(map[string]data.Post)}
}

func (r *testRepo) ReplaceSnapshot(ctx context.Context, snapshot data.Snapshot) error {
	r.posts = make(map[string]data.Post, len(snapshot.Posts))
	for _, post := range snapshot.Posts {
		r.posts[post.ID] = post
	}
	return nil
}

func (r *testRepo) Search(ctx context.Context, query string, limit, offset int) ([]data.SearchRow, error) {
	rows := make([]data.SearchRow, 0)
	for _, post := range r.posts {
		rows = append(rows, data.SearchRow{Post: post, Score: 1})
	}
	return rows, nil
}

func TestRouterIngestAuth(t *testing.T) {
	repo := newTestRepo()
	ingest := service.NewIngestService(repo)
	search := service.NewSearchService(repo, ngram.NewTokenizer(2), snippet.NewBuilder(), 50)
	router := NewRouter(Dependencies{
		IngestService: ingest,
		SearchService: search,
		AuthToken:     "secret",
		Logger:        log.New(),
	})

	payload := map[string]any{
		"snapshot_id":  "snapshot-1",
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"posts": []map[string]string{
			{
				"id":           "post-1",
				"title":        "Title",
				"url":          "https://example.com/1",
				"content":      "hello world",
				"published_at": time.Now().UTC().Format(time.RFC3339),
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/v1/posts/snapshot", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.Code)
	}

	req2 := httptest.NewRequest(http.MethodPut, "/v1/posts/snapshot", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer secret")
	resp2 := httptest.NewRecorder()
	router.ServeHTTP(resp2, req2)
	if resp2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp2.Code)
	}
}

func TestRouterSearch(t *testing.T) {
	repo := newTestRepo()
	now := time.Now().UTC()
	_ = repo.ReplaceSnapshot(context.Background(), data.Snapshot{
		SnapshotID:  "snapshot-for-search",
		GeneratedAt: now,
		Posts: []data.Post{{
			ID:          "post-1",
			Title:       "中文检索实践",
			URL:         "https://example.com/1",
			Content:     "介绍中文搜索实现。",
			PublishedAt: now,
		}},
	})

	router := NewRouter(Dependencies{
		IngestService: service.NewIngestService(repo),
		SearchService: service.NewSearchService(repo, ngram.NewTokenizer(2), snippet.NewBuilder(), 50),
		AuthToken:     "secret",
		Logger:        log.New(),
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/search?q=%E4%B8%AD%E6%96%87", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}
