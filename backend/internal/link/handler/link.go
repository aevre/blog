// Package handler exposes link HTTP endpoints and transport DTOs.
package handler

import (
	linkmodel "blog-server/internal/link/model"
	"blog-server/internal/link/service"
	"blog-server/internal/shared/errx"
	"blog-server/internal/shared/httpx"
	"blog-server/internal/shared/validatorx"

	"github.com/labstack/echo/v5"
)

// LinkHandler defines the interface for link HTTP handlers.
type LinkHandler interface {
	GetLinks(c *echo.Context) error
	ApplyForALinks(c *echo.Context) error
}

// linkHandler implements the LinkHandler interface.
type linkHandler struct {
	svc      service.LinkService
	validate validatorx.Validator
}

// NewLinkHandler creates a new link handler instance.
func NewLinkHandler(svc service.LinkService, validate validatorx.Validator) LinkHandler {
	return &linkHandler{svc: svc, validate: validate}
}

// GetLinks retrieves all enabled links.
func (h *linkHandler) GetLinks(c *echo.Context) error {
	links, err := h.svc.GetLinks(c.Request().Context())
	if err != nil {
		return err
	}

	linkDTOs := make([]LinkRes, len(links))
	for i, link := range links {
		linkDTOs[i] = toLinkResponse(link)
	}

	return httpx.OK(c, httpx.Success(linkDTOs))
}

// ApplyForALinks creates a new link application.
func (h *linkHandler) ApplyForALinks(c *echo.Context) error {
	req := new(CreateLinkReq)
	if err := c.Bind(req); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	if err := h.validate.Struct(req); err != nil {
		return errx.New(errx.CodeValidationFailed, err)
	}

	input := &service.CreateLinkInput{
		Name:        req.Name,
		Description: &req.Description,
		Avatar:      &req.Avatar,
		URL:         req.URL,
	}
	err := h.svc.CreateLink(
		c.Request().Context(),
		input,
	)
	if err != nil {
		return err
	}

	return httpx.OK(c, httpx.Success(""))
}

// RegisterLinkRoutes registers all link-related routes.
func RegisterLinkRoutes(r *echo.Group, h LinkHandler) {
	group := r.Group("/links")
	group.GET("", h.GetLinks)
	group.POST("/apply-link", h.ApplyForALinks)
}

// toLinkResponse maps a domain Link to the response DTO.
func toLinkResponse(link *linkmodel.Link) LinkRes {
	res := LinkRes{
		ID:        link.ID,
		Name:      link.Name,
		URL:       link.URL,
		SortOrder: link.SortOrder,
	}

	if link.Description != nil {
		res.Description = *link.Description
	}
	if link.Avatar != nil {
		res.Avatar = *link.Avatar
	}

	return res
}
