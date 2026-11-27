package handler

import (
	"net/http/httptest"
	"testing"

	accountmodel "blog-server/internal/account/model"
	"blog-server/internal/shared/contextx"
	"blog-server/internal/shared/errx"

	"github.com/labstack/echo/v5"
)

func TestAdminLinkMutationsRequireAdmin(t *testing.T) {
	for _, tc := range []struct {
		name   string
		method string
		handle func(*linkHandler, *echo.Context) error
	}{
		{name: "update", method: "PATCH", handle: (*linkHandler).AdminUpdate},
		{name: "delete", method: "DELETE", handle: (*linkHandler).AdminDelete},
	} {
		for _, role := range []accountmodel.UserRole{accountmodel.UserRoleReader, accountmodel.UserRoleAdmin} {
			t.Run(tc.name+"/"+string(role), func(t *testing.T) {
				e := echo.New()
				c := e.NewContext(httptest.NewRequest(tc.method, "/admin/links/invalid", nil), httptest.NewRecorder())
				c.SetPathValues(echo.PathValues{{Name: "id", Value: "invalid"}})
				c.SetRequest(c.Request().WithContext(contextx.SetUser(c.Request().Context(), contextx.User{ID: 1, Role: role})))
				err := tc.handle(&linkHandler{}, c)
				want := errx.CodeForbidden
				if role == accountmodel.UserRoleAdmin {
					want = errx.CodeInvalidParam // Admin passes the role check and reaches ID validation.
				}
				if err == nil || errx.ToAppError(err).Code != want {
					t.Fatalf("want error code %d, got %v", want, err)
				}
			})
		}
	}
}
