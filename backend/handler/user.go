package handler

import (
	"fmt"

	"blog-server/contextx"
	"blog-server/entity"
	"blog-server/middleware"
	"blog-server/pkg/errx"
	"blog-server/response"
	"blog-server/service"

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

	return response.OK(c, response.Success(toUserRes(user)))
}

func RegisterUserRoutes(r *echo.Group, h UserHandler, auth *middleware.AuthMiddleware) {
	r.GET("/user/me", h.Me, auth.Handler())
}

func toUserRes(user *entity.User) *response.UserRes {
	return &response.UserRes{
		UUID:      user.UUID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
