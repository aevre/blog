package handler

import contentmodel "blog-server/internal/content/model"

// CreatePostReq is the request body for creating a post.
// UserID is intentionally omitted — it is extracted from the JWT context
// by the handler, never accepted from the client.
type CreatePostReq struct {
	Title   string                  `json:"title" validate:"required,min=3,max=100"`
	Summary *string                 `json:"summary"`
	Cover   *string                 `json:"cover"`
	Content string                  `json:"content" validate:"required"`
	Status  contentmodel.PostStatus `json:"status"`

	CategoryIDs []uint `json:"categoryIDs"`
	Tags        []uint `json:"tags"`
}

// UpdatePostReq is the request body for updating a post.
type UpdatePostReq struct {
	Title   *string                  `json:"title" validate:"omitempty,min=3,max=100"`
	Summary *string                  `json:"summary" validate:"omitempty"`
	Cover   *string                  `json:"cover" validate:"omitempty"`
	Content *string                  `json:"content" validate:"omitempty"`
	Status  *contentmodel.PostStatus `json:"status" validate:"omitempty,oneof=draft published archived"`

	CategoryIDs *[]uint `json:"categoryIDs"`
	Tags        *[]uint `json:"tags"`
}

// PostPageReq is the request query for paginating posts.
type PostPageReq struct {
	Page     int `json:"page" query:"page" validate:"omitempty,min=1"`
	PageSize int `json:"pageSize" query:"pageSize" validate:"omitempty,min=1,max=100"`
}

// AdminPostListReq is the request query for admin post list.
type AdminPostListReq struct {
	Page     int                      `json:"page" query:"page" validate:"omitempty,min=1"`
	PageSize int                      `json:"pageSize" query:"pageSize" validate:"omitempty,min=1,max=100"`
	Status   *contentmodel.PostStatus `json:"status" query:"status" validate:"omitempty,oneof=draft published archived"`
	Keyword  *string                  `json:"keyword" query:"keyword" validate:"omitempty,max=100"`
}
