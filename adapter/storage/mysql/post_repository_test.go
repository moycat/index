package mysql

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/moycat/index/data"
)

func TestPostRepositoryReindexAllPosts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostRepository(db)
	now := time.Now().UTC()
	req := data.IndexRequest{
		Posts: []data.Post{{
			Title:       "title",
			URL:         "https://example.com/1",
			Content:     "content",
			PublishedAt: now,
		}},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM posts")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO posts (title, url, content, published_at)")).
		WithArgs(req.Posts[0].Title, req.Posts[0].URL, req.Posts[0].Content, req.Posts[0].PublishedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.ReindexAllPosts(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostRepositorySearch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostRepository(db)
	now := time.Now().UTC()

	rows := sqlmock.NewRows([]string{"title", "url", "content", "published_at", "score"}).
		AddRow("中文", "https://example.com/1", "content", now, 1.2)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT title, url, content, published_at,")).
		WithArgs("中文", "中文", 10, 0).
		WillReturnRows(rows)

	result, err := repo.Search(context.Background(), "中文", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
