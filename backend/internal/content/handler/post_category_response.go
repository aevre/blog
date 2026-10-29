package handler

// PostCategoryRes represents a category associated with a post.
type PostCategoryRes struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
