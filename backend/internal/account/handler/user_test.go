package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	accountmodel "blog-server/internal/account/model"
	"blog-server/internal/account/service"
	"blog-server/internal/shared/contextx"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type userServiceStub struct {
	service.UserService
	user *accountmodel.User
	id   uint
}

func (s *userServiceStub) GetByID(_ context.Context, id uint) (*accountmodel.User, error) {
	s.id = id
	return s.user, nil
}

func TestUserMe(t *testing.T) {
	createdAt := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	user := &accountmodel.User{
		ID:        42,
		UUID:      uuid.New(),
		Username:  "reader",
		Email:     "reader@example.com",
		Role:      accountmodel.UserRoleReader,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
	svc := &userServiceStub{user: user}
	e := echo.New()
	recorder := httptest.NewRecorder()
	ctx := e.NewContext(httptest.NewRequest("GET", "/auth/me", nil), recorder)
	ctx.SetRequest(ctx.Request().WithContext(contextx.SetUser(ctx.Request().Context(), contextx.User{ID: user.ID})))

	if err := (&userHandler{svc: svc}).Me(ctx); err != nil {
		t.Fatal(err)
	}

	if svc.id != user.ID {
		t.Fatalf("queried user ID = %d, want %d", svc.id, user.ID)
	}
	var body struct {
		Code int             `json:"code"`
		Data responseUserRes `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 || body.Data.UUID != user.UUID.String() || body.Data.Email != user.Email {
		t.Fatalf("unexpected response: %+v", body)
	}
}

type responseUserRes struct {
	UUID  string `json:"uuid"`
	Email string `json:"email"`
}

func TestUserMeRequiresAuthenticatedContext(t *testing.T) {
	e := echo.New()
	ctx := e.NewContext(httptest.NewRequest("GET", "/auth/me", nil), httptest.NewRecorder())
	if err := (&userHandler{}).Me(ctx); err == nil {
		t.Fatal("expected unauthorized error without authenticated user")
	}
}
