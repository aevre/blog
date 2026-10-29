package api

import (
	"testing"

	accounthandler "blog-server/internal/account/handler"
	contenthandler "blog-server/internal/content/handler"
	linkhandler "blog-server/internal/link/handler"
	"blog-server/internal/platform/config"
	"blog-server/internal/platform/middleware"
	setuphandler "blog-server/internal/setup/handler"
	summaryhandler "blog-server/internal/summary/handler"

	"github.com/labstack/echo/v5"
)

func TestExistingRoutes(t *testing.T) {
	e := echo.New()
	RegisterRoutes(e, Handlers{
		Auth: accounthandler.NewAuthHandler(nil, nil, nil), User: accounthandler.NewUserHandler(nil),
		Post: contenthandler.NewPostHandler(nil, nil), Rss: contenthandler.NewRssHandler(nil),
		Link: linkhandler.NewLinkHandler(nil, nil), Setup: setuphandler.NewSetupHandler(nil, nil), Model: summaryhandler.NewModelHandler(nil, nil),
	}, middleware.NewAuthMiddleware(&config.Config{}))
	for _, r := range []struct{ method, path string }{
		{"GET", "/setup"},
		{"POST", "/setup"},
		{"POST", "/auth/captcha"},
		{"POST", "/auth/register"},
		{"POST", "/auth/login"},
		{"POST", "/auth/logout"},
		{"POST", "/auth/refresh"},
		{"GET", "/user/me"},
		{"GET", "/posts"},
		{"GET", "/posts/meta"},
		{"GET", "/posts/:id"},
		{"POST", "/posts"},
		{"GET", "/admin/posts"},
		{"GET", "/admin/posts/:id"},
		{"PUT", "/admin/posts/:id"},
		{"DELETE", "/admin/posts/:id"},
		{"GET", "/rss"},
		{"GET", "/rss/complete"},
		{"GET", "/links"},
		{"POST", "/links/apply-link"},
		{"POST", "/model/summarize"},
		{"GET", "/model/summarize/:sessionId"},
	} {
		if _, err := e.Router().Routes().FindByMethodPath(r.method, "/api/v1"+r.path); err != nil {
			t.Errorf("missing %s %s: %v", r.method, r.path, err)
		}
	}
}
