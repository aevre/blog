package handler

import (
	accountmodel "blog-server/internal/account/model"
	"blog-server/internal/authz"
	linkmodel "blog-server/internal/link/model"
	"blog-server/internal/link/service"
	"blog-server/internal/shared/errx"
	"blog-server/internal/shared/httpx"

	"github.com/labstack/echo/v5"
)

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
