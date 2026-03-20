package data

import (
	"testing"
	"time"
)

func TestIndexRequestValidate(t *testing.T) {
	now := time.Now().UTC()
	valid := IndexRequest{
		Posts: []Post{{
			Title:       "title",
			URL:         "https://example.com/1",
			Content:     "content",
			PublishedAt: now,
		}},
	}

	tests := []struct {
		name    string
		req     IndexRequest
		wantErr bool
	}{
		{name: "valid", req: valid},
		{name: "invalid post", req: IndexRequest{Posts: []Post{{}}}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
