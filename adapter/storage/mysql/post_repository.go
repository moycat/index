package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/moycat/index/data"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) ReplaceSnapshot(ctx context.Context, snapshot data.Snapshot) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var exists int
	err = tx.QueryRowContext(ctx, "SELECT 1 FROM ingest_snapshots WHERE snapshot_id = ? LIMIT 1", snapshot.SnapshotID).Scan(&exists)
	if err == nil {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit idempotent snapshot transaction: %w", err)
		}
		return nil
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("check snapshot idempotency: %w", err)
	}

	const upsertStmt = `
INSERT INTO posts (id, title, url, content, published_at, snapshot_id)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	title = VALUES(title),
	url = VALUES(url),
	content = VALUES(content),
	published_at = VALUES(published_at),
	snapshot_id = VALUES(snapshot_id)
`
	for _, post := range snapshot.Posts {
		if _, err := tx.ExecContext(ctx, upsertStmt,
			post.ID,
			post.Title,
			post.URL,
			post.Content,
			post.PublishedAt,
			snapshot.SnapshotID,
		); err != nil {
			return fmt.Errorf("upsert post in snapshot: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM posts WHERE snapshot_id <> ?", snapshot.SnapshotID); err != nil {
		return fmt.Errorf("delete stale posts: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		"INSERT INTO ingest_snapshots (snapshot_id, generated_at, post_count) VALUES (?, ?, ?)",
		snapshot.SnapshotID,
		snapshot.GeneratedAt,
		len(snapshot.Posts),
	); err != nil {
		return fmt.Errorf("insert snapshot record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit snapshot transaction: %w", err)
	}
	return nil
}

func (r *PostRepository) Search(ctx context.Context, query string, limit, offset int) ([]data.SearchRow, error) {
	const stmt = `
SELECT id, title, url, content, published_at,
       MATCH(title, content) AGAINST (? IN NATURAL LANGUAGE MODE) AS score
FROM posts
WHERE MATCH(title, content) AGAINST (? IN NATURAL LANGUAGE MODE)
ORDER BY score DESC, published_at DESC
LIMIT ? OFFSET ?
`
	rows, err := r.db.QueryContext(ctx, stmt, query, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query search posts: %w", err)
	}
	defer rows.Close()

	results := make([]data.SearchRow, 0, limit)
	for rows.Next() {
		var row data.SearchRow
		if err := rows.Scan(
			&row.Post.ID,
			&row.Post.Title,
			&row.Post.URL,
			&row.Post.Content,
			&row.Post.PublishedAt,
			&row.Score,
		); err != nil {
			return nil, fmt.Errorf("scan search row: %w", err)
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search rows: %w", err)
	}
	return results, nil
}
