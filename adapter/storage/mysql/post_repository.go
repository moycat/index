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

func (r *PostRepository) ReindexAllPosts(ctx context.Context, req data.IndexRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM posts"); err != nil {
		return fmt.Errorf("delete existing posts: %w", err)
	}

	const insertStmt = `
INSERT INTO posts (title, url, content, published_at)
VALUES (?, ?, ?, ?)
`
	for _, post := range req.Posts {
		if _, err := tx.ExecContext(ctx, insertStmt,
			post.Title,
			post.URL,
			post.Content,
			post.PublishedAt,
		); err != nil {
			return fmt.Errorf("insert post in index request: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit reindex transaction: %w", err)
	}
	return nil
}

func (r *PostRepository) Search(ctx context.Context, query string, limit, offset int) ([]data.SearchRow, error) {
	const stmt = `
SELECT title, url, content, published_at,
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
