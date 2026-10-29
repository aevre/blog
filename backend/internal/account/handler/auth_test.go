package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"blog-server/internal/account/service"
	"blog-server/internal/platform/config"
	"blog-server/internal/shared/validatorx"

	"github.com/labstack/echo/v5"
)

type loginServiceStub struct{ service.AuthService }

func (*loginServiceStub) Login(context.Context, *service.LoginInput) (*service.AuthResult, error) {
	return &service.AuthResult{AccessToken: "access", RefreshToken: "refresh", UUID: "id", Username: "reader", Role: "reader"}, nil
}

func TestLoginContract(t *testing.T) {
	cfg := &config.Config{App: config.AppConfig{Environment: config.EnvProd}, JWT: config.JWTConfig{AccessExpiration: time.Minute, RefreshExpiration: time.Hour}}
	h := NewAuthHandler(&loginServiceStub{}, validatorx.NewValidator(), cfg)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{"email":"reader@example.com","password":"Password123!"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e := echo.New()
	if err := h.Login(e.NewContext(req, rec)); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 {
		t.Fatalf("status=%d", rec.Code)
	}
	var body struct {
		Code int            `json:"code"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"accessToken", "uuid", "avatar", "username", "role"} {
		if _, ok := body.Data[field]; !ok {
			t.Errorf("missing field %s: %s", field, rec.Body.String())
		}
	}
	if _, ok := body.Data["refreshToken"]; ok {
		t.Fatal("refresh token exposed in JSON")
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("cookies=%v", cookies)
	}
	for _, cookie := range cookies {
		if !cookie.HttpOnly || !cookie.Secure || cookie.Path != "/" {
			t.Errorf("cookie=%+v", cookie)
		}
	}
}
