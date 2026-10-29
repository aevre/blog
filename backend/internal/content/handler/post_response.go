package handler

import "time"

// PostListRes is the response model for the public post list displayed on the frontend.
type PostListRes struct {
	ID              uint              `json:"id"`
	Title           string            `json:"title"`
	Cover           *string           `json:"cover"`
	Summary         *string           `json:"summary"`
	ReadTimeMinutes uint              `json:"readTimeMinutes"`
	ViewCount       uint              `json:"viewCount"`
	PublishedAt     *time.Time        `json:"publishedAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	Author          string            `json:"author"`
	Tags            []PostTagRes      `json:"tags"`
	Categories      []PostCategoryRes `json:"categories"`
}

// PostRes represents the response for a public post.
type PostRes struct {
	ID              uint              `json:"id"`
	Title           string            `json:"title"`
	Summary         *string           `json:"summary"`
	Content         string            `json:"content"`
	Cover           *string           `json:"cover"`
	ReadTimeMinutes uint              `json:"readTimeMinutes"`
	ViewCount       uint              `json:"viewCount"`
	PublishedAt     *time.Time        `json:"publishedAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	Author          string            `json:"author"`
	Tags            []PostTagRes      `json:"tags"`
	Categories      []PostCategoryRes `json:"categories"`
}

// PostMetaRes represents post metadata used for sitemap generation.
type PostMetaRes struct {
	ID        uint      `json:"id"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AdminPostListRes represents the post list response for the admin panel.
type AdminPostListRes struct {
	ID              uint              `json:"id"`
	Title           string            `json:"title"`
	Cover           *string           `json:"cover"`
	Summary         *string           `json:"summary"`
	Status          string            `json:"status"`
	ReadTimeMinutes uint              `json:"readTimeMinutes"`
	ViewCount       uint              `json:"viewCount"`
	PublishedAt     *time.Time        `json:"publishedAt"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	Author          string            `json:"author"`
	Tags            []PostTagRes      `json:"tags"`
	Categories      []PostCategoryRes `json:"categories"`
}

// AdminPostRes represents the post detail response for the admin panel.
type AdminPostRes struct {
	ID              uint              `json:"id"`
	Title           string            `json:"title"`
	Summary         *string           `json:"summary"`
	Content         string            `json:"content"`
	Cover           *string           `json:"cover"`
	Status          string            `json:"status"`
	ReadTimeMinutes uint              `json:"readTimeMinutes"`
	ViewCount       uint              `json:"viewCount"`
	PublishedAt     *time.Time        `json:"publishedAt"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	Author          string            `json:"author"`
	Tags            []PostTagRes      `json:"tags"`
	Categories      []PostCategoryRes `json:"categories"`
}
