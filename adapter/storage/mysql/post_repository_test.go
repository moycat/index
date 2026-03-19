package mysql

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/moycat/index/data"
)

func TestPostRepositoryReplaceSnapshot(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostRepository(db)
	now := time.Now().UTC()
	snapshot := data.Snapshot{
		SnapshotID:  "snapshot-1",
		GeneratedAt: now,
		Posts: []data.Post{{
			ID:          "post-1",
			Title:       "title",
			URL:         "https://example.com/1",
			Content:     "content",
			PublishedAt: now,
		}},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM ingest_snapshots WHERE snapshot_id = ? LIMIT 1")).
		WithArgs(snapshot.SnapshotID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO posts (id, title, url, content, published_at, snapshot_id)")).
		WithArgs(snapshot.Posts[0].ID, snapshot.Posts[0].Title, snapshot.Posts[0].URL, snapshot.Posts[0].Content, snapshot.Posts[0].PublishedAt, snapshot.SnapshotID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM posts WHERE snapshot_id <> ?")).
		WithArgs(snapshot.SnapshotID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ingest_snapshots (snapshot_id, generated_at, post_count) VALUES (?, ?, ?)")).
		WithArgs(snapshot.SnapshotID, snapshot.GeneratedAt, len(snapshot.Posts)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.ReplaceSnapshot(context.Background(), snapshot); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostRepositoryReplaceSnapshotIdempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostRepository(db)
	now := time.Now().UTC()
	snapshot := data.Snapshot{SnapshotID: "snapshot-existing", GeneratedAt: now}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM ingest_snapshots WHERE snapshot_id = ? LIMIT 1")).
		WithArgs(snapshot.SnapshotID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))
	mock.ExpectCommit()

	if err := repo.ReplaceSnapshot(context.Background(), snapshot); err != nil {
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

	rows := sqlmock.NewRows([]string{"id", "title", "url", "content", "published_at", "score"}).
		AddRow("post-1", "中文", "https://example.com/1", "content", now, 1.2)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, url, content, published_at,")).
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
