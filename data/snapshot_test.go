package data

import (
	"testing"
	"time"
)

func TestSnapshotValidate(t *testing.T) {
	now := time.Now().UTC()
	valid := Snapshot{
		SnapshotID:  "snapshot-1",
		GeneratedAt: now,
		Posts: []Post{{
			ID:          "post-1",
			Title:       "title",
			URL:         "https://example.com/1",
			Content:     "content",
			PublishedAt: now,
		}},
	}

	tests := []struct {
		name    string
		snap    Snapshot
		wantErr bool
	}{
		{name: "valid", snap: valid},
		{name: "missing snapshot id", snap: Snapshot{}, wantErr: true},
		{name: "missing generated_at", snap: Snapshot{SnapshotID: "x"}, wantErr: true},
		{name: "duplicate post ids", snap: Snapshot{SnapshotID: "x", GeneratedAt: now, Posts: []Post{valid.Posts[0], valid.Posts[0]}}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.snap.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
