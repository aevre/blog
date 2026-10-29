package handler

import (
	"fmt"

	accountmodel "blog-server/internal/account/model"
	"blog-server/internal/account/service"
	"blog-server/internal/platform/middleware"
	"blog-server/internal/shared/contextx"
	"blog-server/internal/shared/errx"
	"blog-server/internal/shared/httpx"

	"github.com/labstack/echo/v5"
)

type UserHandler interface {
	Me(c *echo.Context) error
}

type userHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) UserHandler {
	return &userHandler{svc: svc}
}

func (h *userHandler) Me(c *echo.Context) error {
	currentUser, ok := contextx.GetUser(c.Request().Context())
	if !ok {
		return errx.New(errx.CodeUnauthorized, fmt.Errorf("missing user in context"))
	}

	user, err := h.svc.GetByID(c.Request().Context(), currentUser.ID)
	if err != nil {
		return err
	}

	return httpx.OK(c, httpx.Success(toUserRes(user)))
}

func RegisterUserRoutes(r *echo.Group, h UserHandler, auth *middleware.AuthMiddleware) {
	r.GET("/user/me", h.Me, auth.Handler())
}

func toUserRes(user *accountmodel.User) *UserRes {
	return &UserRes{
		UUID:      user.UUID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
