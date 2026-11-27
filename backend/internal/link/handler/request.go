package handler

import "blog-server/internal/platform/field"

type CreateLinkReq struct {
	Name        string  `json:"name" validate:"required"`
	URL         string  `json:"url" validate:"required,url"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	Avatar      *string `json:"avatar" validate:"omitempty,url,max=255"`
	CategoryID  *uint   `json:"CategoryID"`
}

type UpdateLinkReq struct {
	Version     int64                  `json:"version" validate:"required,min=1"`
	Name        field.Optional[string] `json:"name" validate:"omitempty,min=1,max=100"`
	URL         field.Optional[string] `json:"url" validate:"omitempty,url"`
	Description field.Optional[string] `json:"description" validate:"omitempty,max=1000"`
	Avatar      field.Optional[string] `json:"avatar" validate:"omitempty,url,max=255"`
	SortOrder   field.Optional[int]    `json:"sortOrder" validate:"omitempty,min=0"`
	CategoryID  field.Optional[uint]   `json:"categoryId" validate:"omitempty,min=1"`
}

type ReviewLinkReq struct {
	Version  int64  `json:"version" validate:"required,min=1"`
	Decision string `json:"decision" validate:"required,oneof=approved rejected"`
	Reason   string `json:"reason" validate:"max=1000"`
}
