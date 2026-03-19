package data

import (
	"fmt"
	"strings"
	"time"
)

// Snapshot is a full replacement payload for the indexed post set.
type Snapshot struct {
	SnapshotID  string
	GeneratedAt time.Time
	Posts       []Post
}

func (s Snapshot) Validate() error {
	if strings.TrimSpace(s.SnapshotID) == "" {
		return fmt.Errorf("snapshot_id is required")
	}
	if s.GeneratedAt.IsZero() {
		return fmt.Errorf("generated_at is required")
	}

	seen := make(map[string]struct{}, len(s.Posts))
	for i, post := range s.Posts {
		if err := post.Validate(); err != nil {
			return fmt.Errorf("posts[%d]: %w", i, err)
		}
		if _, ok := seen[post.ID]; ok {
			return fmt.Errorf("duplicate post id: %s", post.ID)
		}
		seen[post.ID] = struct{}{}
	}
	return nil
}
