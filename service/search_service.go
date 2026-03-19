package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/moycat/index/data"
)

const (
	defaultPageSize = 10
	maxPageSize     = 50
)

type SearchService struct {
	repo          PostRepository
	tokenizer     Tokenizer
	snippet       SnippetBuilder
	snippetRunes  int
	defaultPaging int
	maxPaging     int
}

func NewSearchService(repo PostRepository, tokenizer Tokenizer, snippet SnippetBuilder, snippetRunes int) *SearchService {
	return &SearchService{
		repo:          repo,
		tokenizer:     tokenizer,
		snippet:       snippet,
		snippetRunes:  snippetRunes,
		defaultPaging: defaultPageSize,
		maxPaging:     maxPageSize,
	}
}

func (s *SearchService) Search(ctx context.Context, query string, page, pageSize int) ([]data.SearchHit, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("%w: query is required", ErrInvalidArgument)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = s.defaultPaging
	}
	if pageSize > s.maxPaging {
		pageSize = s.maxPaging
	}

	offset := (page - 1) * pageSize
	rows, err := s.repo.Search(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("search posts: %w", err)
	}

	terms := s.tokenizer.Tokenize(query)
	hits := make([]data.SearchHit, 0, len(rows))
	for _, row := range rows {
		matchedTerms := matchedTerms(terms, row.Post.Title+"\n"+row.Post.Content)
		hits = append(hits, data.SearchHit{
			Title:        row.Post.Title,
			URL:          row.Post.URL,
			Snippet:      s.snippet.Build(row.Post.Content, matchedTerms, s.snippetRunes),
			Score:        row.Score,
			MatchedTerms: matchedTerms,
		})
	}

	return hits, nil
}

func matchedTerms(tokens []string, text string) []string {
	if len(tokens) == 0 {
		return nil
	}
	lowerText := strings.ToLower(text)
	seen := make(map[string]struct{}, len(tokens))
	terms := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		if strings.Contains(lowerText, strings.ToLower(token)) {
			seen[token] = struct{}{}
			terms = append(terms, token)
		}
	}
	return terms
}
