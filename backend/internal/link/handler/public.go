// Package handler exposes link HTTP endpoints and transport DTOs.
package handler

import (
	"errors"
	"strconv"

	"blog-server/internal/link/service"
	"blog-server/internal/platform/middleware"
	"blog-server/internal/shared/contextx"
	"blog-server/internal/shared/errx"
	"blog-server/internal/shared/httpx"
	"blog-server/internal/shared/validatorx"

	"github.com/labstack/echo/v5"
)

// LinkHandler defines the interface for link HTTP handlers.
type LinkHandler interface {
	ListPublic(c *echo.Context) error

	AdminList(c *echo.Context) error
	AdminUpdate(c *echo.Context) error
	AdminDelete(c *echo.Context) error
	Create(c *echo.Context) error
	Update(c *echo.Context) error
	Delete(c *echo.Context) error
	Review(c *echo.Context) error

	GetMineByID(c *echo.Context) error
	AdminGetByID(c *echo.Context) error
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

func requestActor(c *echo.Context) (contextx.User, error) {
	actor, ok := contextx.GetUser(c.Request().Context())
	if !ok || actor.ID == 0 {
		return contextx.User{}, errx.New(errx.CodeUnauthorized, errors.New("authentication required"))
	}
	return actor, nil
}

func requestLinkID(c *echo.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, errx.New(errx.CodeInvalidParam, errors.New("invalid link ID"))
	}
	return uint(id), nil
}

// ListPublic retrieves all enabled links.
func (h *linkHandler) ListPublic(c *echo.Context) error {
	links, err := h.svc.ListPublic(c.Request().Context())
	if err != nil {
		return err
	}

	linkDTOs := toPublicLinks(links)

	return httpx.OK(c, httpx.Success(linkDTOs))
}

func (h *linkHandler) Create(c *echo.Context) error {
	actor, err := requestActor(c)
	if err != nil {
		return err
	}

	req := new(CreateLinkReq)
	if err := c.Bind(req); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	if err := h.validate.Struct(req); err != nil {
		return errx.New(errx.CodeValidationFailed, err)
	}

	created, err := h.svc.Create(c.Request().Context(), actor, service.CreateInput{
		Name:        req.Name,
		URL:         req.URL,
		Description: req.Description,
		Avatar:      req.Avatar,
		CategoryID:  req.CategoryID,
	})
	if err != nil {
		return err
	}

	return httpx.OK(c, httpx.Success(
		toManagedLink(created),
	))
}

func (h *linkHandler) Update(c *echo.Context) error {
	actor, err := requestActor(c)
	if err != nil {
		return err
	}
	id, err := requestLinkID(c)
	if err != nil {
		return err
	}
	req := new(UpdateLinkReq)
	if err := c.Bind(req); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}
	if err := h.validate.Struct(req); err != nil {
		return errx.New(errx.CodeValidationFailed, err)
	}
	updated, err := h.svc.Update(c.Request().Context(), actor, service.UpdateInput{
		ID:          id,
		Version:     req.Version,
		Name:        req.Name,
		URL:         req.URL,
		Description: req.Description,
		Avatar:      req.Avatar,
		SortOrder:   req.SortOrder,
		CategoryID:  req.CategoryID,
	})
	if err != nil {
		return err
	}

	return httpx.OK(c, httpx.Success(
		toManagedLink(updated),
	))
}

func (h *linkHandler) Delete(c *echo.Context) error {
	actor, err := requestActor(c)
	if err != nil {
		return err
	}
	id, err := requestLinkID(c)
	if err != nil {
		return err
	}
	req := new(struct {
		Version int64 `json:"version" validate:"required,min=1"`
	})
	if err := c.Bind(req); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}
	if err := h.validate.Struct(req); err != nil {
		return errx.New(errx.CodeValidationFailed, err)
	}
	err = h.svc.Delete(c.Request().Context(), actor, id, req.Version)
	if err != nil {
		return err
	}

	return httpx.OK(c, httpx.Success[any](nil))
}

func (h *linkHandler) GetMineByID(c *echo.Context) error {
	actor, err := requestActor(c)
	if err != nil {
		return err
	}
	id, err := requestLinkID(c)
	if err != nil {
		return err
	}
	l, err := h.svc.GetMineByID(c.Request().Context(), actor, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, httpx.Success(toManagedLink(l)))
}

// RegisterLinkRoutes registers all link-related routes.
func RegisterLinkRoutes(r *echo.Group, h LinkHandler, auth *middleware.AuthMiddleware) {
	public := r.Group("/links")
	public.GET("", h.ListPublic)

	signed := r.Group("/links", auth.Handler())
	signed.POST("", h.Create)
	signed.GET("/mine/:id", h.GetMineByID)
	signed.PATCH("/mine/:id", h.Update)
	signed.DELETE("/mine/:id", h.Delete)

	admin := r.Group("/admin/links", auth.Handler())
	admin.GET("", h.AdminList)
	admin.GET("/:id", h.AdminGetByID)
	admin.PATCH("/:id", h.AdminUpdate)
	admin.DELETE("/:id", h.AdminDelete)
	admin.POST("/:id/review", h.Review)
}
