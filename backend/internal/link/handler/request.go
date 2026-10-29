package handler

type CreateLinkReq struct {
	Name        string `json:"name" validate:"required"`
	URL         string `json:"url" validate:"required,url"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
}
