package handler

// PostTagRes represents a tag associated with a post.
type PostTagRes struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
