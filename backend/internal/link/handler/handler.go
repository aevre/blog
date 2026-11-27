// Package handler exposes link HTTP endpoints and transport DTOs.
package handler

import (
	"errors"
	"strconv"

	accountmodel "blog-server/internal/account/model"
	"blog-server/internal/authz"
	linkmodel "blog-server/internal/link/model"
	"blog-server/internal/link/service"
	"blog-server/internal/shared/contextx"
	"blog-server/internal/shared/errx"
	"blog-server/internal/shared/httpx"
	"blog-server/internal/shared/validatorx"

	"github.com/labstack/echo/v5"
)

// LinkHandler defines the interface for link HTTP handlers.
type LinkHandler interface {
	ListPublic(c *echo.Context) error

	Create(c *echo.Context) error
	GetMineByID(c *echo.Context) error
	Update(c *echo.Context) error
	Delete(c *echo.Context) error

	AdminList(c *echo.Context) error
	AdminGetByID(c *echo.Context) error
	AdminUpdate(c *echo.Context) error
	AdminDelete(c *echo.Context) error
	Review(c *echo.Context) error
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

// ListPublic returns approved links without requiring authentication.
func (h *linkHandler) ListPublic(c *echo.Context) error {
	links, err := h.svc.ListPublic(c.Request().Context())
	if err != nil {
		return err
	}

	linkDTOs := toPublicLinks(links)

	return httpx.OK(c, httpx.Success(linkDTOs))
}

// Create submits a link for the authenticated user.
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

// GetMineByID returns a link submitted by the authenticated user.
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

// Update handles both owner and admin edits; the service enforces the resource scope.
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

// Delete handles both owner and admin deletions using the supplied version.
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

// AdminList returns all non-deleted links for administrators.
func (h *linkHandler) AdminList(c *echo.Context) error {
	actor, err := requestActor(c)
	if err != nil {
		return err
	}

	links, err := h.svc.AdminList(c.Request().Context(), actor)
	if err != nil {
		return err
	}
	return httpx.OK(c, httpx.Success(toManagedLinks(links)))
}

// AdminGetByID returns a link regardless of its submitter.
func (h *linkHandler) AdminGetByID(c *echo.Context) error {
	actor, err := requestActor(c)
	if err != nil {
		return err
	}
	id, err := requestLinkID(c)
	if err != nil {
		return err
	}
	l, err := h.svc.AdminGetByID(c.Request().Context(), actor, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, httpx.Success(toManagedLink(l)))
}

// AdminUpdate checks the admin role before reusing Update.
func (h *linkHandler) AdminUpdate(c *echo.Context) error {
	actor, err := requestActor(c)
	if err != nil {
		return err
	}
	if actor.Role != accountmodel.UserRoleAdmin {
		return authz.ErrForbidden()
	}
	return h.Update(c)
}

// AdminDelete checks the admin role before reusing Delete.
func (h *linkHandler) AdminDelete(c *echo.Context) error {
	actor, err := requestActor(c)
	if err != nil {
		return err
	}
	if actor.Role != accountmodel.UserRoleAdmin {
		return authz.ErrForbidden()
	}
	return h.Delete(c)
}

// Review submits an approval or rejection decision for a link.
func (h *linkHandler) Review(c *echo.Context) error {
	actor, err := requestActor(c)
	if err != nil {
		return err
	}
	id, err := requestLinkID(c)
	if err != nil {
		return err
	}

	req := new(ReviewLinkReq)
	if err := c.Bind(req); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}
	if err := h.validate.Struct(req); err != nil {
		return errx.New(errx.CodeValidationFailed, err)
	}

	err = h.svc.Review(c.Request().Context(), actor, service.ReviewInput{
		ID:       id,
		Version:  req.Version,
		Decision: linkmodel.LinkReviewStatus(req.Decision),
		Reason:   &req.Reason,
	})
	if err != nil {
		return err
	}
	return httpx.OK(c, httpx.Success[any](nil))
}
