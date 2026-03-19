package data

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Post is the canonical blog document stored and searched by the service.
type Post struct {
	ID          string
	Title       string
	URL         string
	Content     string
	PublishedAt time.Time
}

func (p Post) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return fmt.Errorf("id is required")
	}
	if strings.TrimSpace(p.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if strings.TrimSpace(p.URL) == "" {
		return fmt.Errorf("url is required")
	}
	if _, err := url.ParseRequestURI(p.URL); err != nil {
		return fmt.Errorf("url is invalid: %w", err)
	}
	if strings.TrimSpace(p.Content) == "" {
		return fmt.Errorf("content is required")
	}
	if p.PublishedAt.IsZero() {
		return fmt.Errorf("published_at is required")
	}
	return nil
}

type SearchRow struct {
	Post  Post
	Score float64
}

type SearchHit struct {
	Title        string   `json:"title"`
	URL          string   `json:"url"`
	Snippet      string   `json:"snippet"`
	Score        float64  `json:"score,omitempty"`
	MatchedTerms []string `json:"matched_terms,omitempty"`
}
