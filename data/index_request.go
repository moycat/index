package data

import (
	"fmt"
)

// IndexRequest is a full payload used to reindex all posts.
type IndexRequest struct {
	Posts []Post
}

func (r IndexRequest) Validate() error {
	for i, post := range r.Posts {
		if err := post.Validate(); err != nil {
			return fmt.Errorf("posts[%d]: %w", i, err)
		}
	}
	return nil
}
