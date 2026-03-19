package data

import (
	"testing"
	"time"
)

func TestPostValidate(t *testing.T) {
	now := time.Now().UTC()
	valid := Post{
		ID:          "post-1",
		Title:       "title",
		URL:         "https://example.com/1",
		Content:     "content",
		PublishedAt: now,
	}

	tests := []struct {
		name    string
		post    Post
		wantErr bool
	}{
		{name: "valid", post: valid},
		{name: "missing id", post: Post{}, wantErr: true},
		{name: "invalid url", post: func() Post { p := valid; p.URL = "bad"; return p }(), wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.post.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
