package handler

import (
	"net/http"

	contentmodel "blog-server/internal/content/model"
	"blog-server/internal/content/service"
	"blog-server/internal/shared/errx"

	"github.com/labstack/echo/v5"
)

// RssHandler defines the interface for RSS HTTP handlers.
type RssHandler interface {
	Subscript(c *echo.Context) error
	Complete(c *echo.Context) error
}

// rssHandler implements the RssHandler interface.
type rssHandler struct {
	svc service.RssService
}

// NewRssHandler creates a new RSS handler instance.
func NewRssHandler(svc service.RssService) RssHandler {
	return &rssHandler{svc: svc}
}

// Subscript handles RSS feed subscription requests.
func (h *rssHandler) Subscript(c *echo.Context) error {
	req := new(RssSubscriptReq)

	if err := c.Bind(req); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	var data *contentmodel.RSS
	var err error

	if req.Page <= 0 {
		data, err = h.svc.GenerateRSSFeed(c.Request().Context())
	} else {
		defaultPageSize := 10
		data, err = h.svc.GeneratePagedFeed(c.Request().Context(), req.Page, defaultPageSize)
	}

	if err != nil {
		return err
	}

	return c.XML(http.StatusOK, data)
}

// Complete handles complete RSS feed requests.
func (h *rssHandler) Complete(c *echo.Context) error {
	data, err := h.svc.GenerateCompleteFeed(c.Request().Context())
	if err != nil {
		return err
	}

	return c.XML(http.StatusOK, data)
}

// RegisterRssRoutes registers all RSS-related routes.
func RegisterRssRoutes(r *echo.Group, h RssHandler) {
	group := r.Group("/rss")
	group.GET("", h.Subscript)
	group.GET("/complete", h.Complete)
}
