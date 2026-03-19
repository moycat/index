package service

import (
	"context"

	"github.com/moycat/index/data"
)

type PostRepository interface {
	ReplaceSnapshot(ctx context.Context, snapshot data.Snapshot) error
	Search(ctx context.Context, query string, limit, offset int) ([]data.SearchRow, error)
}

type Tokenizer interface {
	Tokenize(text string) []string
}

type SnippetBuilder interface {
	Build(content string, terms []string, maxRunes int) string
}
