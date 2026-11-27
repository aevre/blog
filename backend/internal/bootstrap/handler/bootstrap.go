// Package handler exposes initial HTTP endpoints and DTOs.
package handler

import (
	"blog-server/internal/bootstrap/service"
	"blog-server/internal/shared/errx"
	"blog-server/internal/shared/httpx"
	"blog-server/internal/shared/validatorx"

	"github.com/labstack/echo/v5"
)

// BootstrapHandler exposes the initial setup status and initialization endpoints.
type BootstrapHandler interface {
	Status(c *echo.Context) error
	Initialize(c *echo.Context) error
}

type bootstrapHandler struct {
	svc      service.BootstrapService
	validate validatorx.Validator
}

// NewBootstrapHandler creates the handler for initial setup.
func NewBootstrapHandler(svc service.BootstrapService, validate validatorx.Validator) BootstrapHandler {
	return &bootstrapHandler{
		svc, validate,
	}
}

// Status reports whether initial setup has completed.
func (h *bootstrapHandler) Status(c *echo.Context) error {
	initialized, err := h.svc.Status(c.Request().Context())
	if err != nil {
		return err
	}

	return httpx.OK(c, struct {
		Initialized bool `json:"initialized"`
	}{
		Initialized: initialized,
	})
}

// Initialize submits the one-time initial setup request.
func (h *bootstrapHandler) Initialize(c *echo.Context) error {
	req := new(InitializeReq)
	if err := c.Bind(req); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	if err := h.validate.Struct(req); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	input := &service.InitializeInput{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		SiteName:    req.SiteName,
		Logo:        req.Logo,
		Greeting:    req.Greeting,
		Description: req.Description,
		Token:       req.Token,
	}

	err := h.svc.Initialize(c.Request().Context(), input)
	if err != nil {
		return err
	}

	return httpx.OK(c, struct {
		Initialized bool `json:"initialized"`
	}{
		Initialized: true,
	})
}

// RegisterBootstrapRouter registers the setup status and initialization routes.
func RegisterBootstrapRouter(r *echo.Group, h BootstrapHandler) {
	group := r.Group("/bootstrap")
	group.GET("", h.Status)
	group.POST("", h.Initialize)
}
