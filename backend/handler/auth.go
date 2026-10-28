package handler

import (
	"fmt"
	"net/http"
	"time"

	"blog-server/config"
	"blog-server/pkg/errx"
	"blog-server/pkg/validatorx"
	"blog-server/request"
	"blog-server/response"
	"blog-server/service"

	"github.com/labstack/echo/v5"
)

// AuthHandler defines the interface for authentication HTTP handlers.
type AuthHandler interface {
	SendCaptcha(c *echo.Context) error
	Register(c *echo.Context) error
	Login(c *echo.Context) error
	Logout(c *echo.Context) error
	Refresh(c *echo.Context) error
}

// authHandler implements the AuthHandler interface.
type authHandler struct {
	svc      service.AuthService
	validate validatorx.Validator
	cfg      *config.Config
}

// NewAuthHandler creates a new auth handler instance.
func NewAuthHandler(authService service.AuthService, validate validatorx.Validator, cfg *config.Config) AuthHandler {
	return &authHandler{svc: authService, validate: validate, cfg: cfg}
}

// SendCaptcha handles sending captcha email for verification.
func (h *authHandler) SendCaptcha(c *echo.Context) error {
	req := new(request.GetCaptchaReq)

	if err := c.Bind(req); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	if err := h.validate.Struct(req); err != nil {
		return errx.New(errx.CodeValidationFailed, err)
	}

	if req.Type == "" {
		req.Type = string(service.Register)
	}

	err := h.svc.SendCaptchaMail(c.Request().Context(), req.Email, service.CaptchaType(req.Type))
	if err != nil {
		return err
	}

	return response.OK(c, response.SuccessWithMsg("Captcha sent successfully", "Captcha sent successfully"))
}

// Register handles user registration.
func (h *authHandler) Register(c *echo.Context) error {
	req := new(request.RegisterReq)

	if err := c.Bind(req); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	if err := h.validate.Struct(req); err != nil {
		return errx.New(errx.CodeValidationFailed, err)
	}

	result, err := h.svc.Register(
		c.Request().Context(),
		&service.RegisterInput{
			Email:    req.Email,
			Password: req.Password,
			Captcha:  req.Captcha,
		},
	)
	if err != nil {
		return err
	}

	h.setRefreshTokenCookie(c, result.RefreshToken)
	h.setAccessTokenCookie(c, result.AccessToken)

	return response.OK(c, response.Success(toLoginRes(result)))
}

// Login handles user login.
func (h *authHandler) Login(c *echo.Context) error {
	req := new(request.LoginReq)

	if err := c.Bind(req); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	if err := h.validate.Struct(req); err != nil {
		return errx.New(errx.CodeValidationFailed, err)
	}

	result, err := h.svc.Login(c.Request().Context(), &service.LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		return err
	}

	h.setRefreshTokenCookie(c, result.RefreshToken)
	h.setAccessTokenCookie(c, result.AccessToken)

	return response.OK(c, response.Success(toLoginRes(result)))
}

// Logout handles user logout by clearing the refresh token cookie.
func (h *authHandler) Logout(c *echo.Context) error {
	if err := h.logoutRefreshToken(c); err != nil {
		return err
	}

	h.clearTokenCookie(c, "refreshToken")
	h.clearTokenCookie(c, "accessToken")

	return response.OK(c, response.SuccessWithMsg("logout success", "logout success"))
}

func (h *authHandler) logoutRefreshToken(c *echo.Context) error {
	token, err := c.Cookie("refreshToken")
	if err != nil || token.Value == "" {
		return nil
	}

	err = h.svc.Logout(c.Request().Context(), token.Value)
	if err == nil {
		return nil
	}

	code := errx.ToAppError(err).Code
	if code == errx.CodeUnauthorized || code == errx.CodeTokenExpired {
		return nil
	}
	return err
}

// Refresh handles access token refresh using refresh token from cookie.
func (h *authHandler) Refresh(c *echo.Context) error {
	token, err := c.Cookie("refreshToken")
	if err != nil {
		return errx.New(errx.CodeUnauthorized, fmt.Errorf("missing refresh token"))
	}
	if token.Value == "" {
		return errx.New(errx.CodeUnauthorized, fmt.Errorf("missing refresh token"))
	}
	accessToken, refreshToken, err := h.svc.RefreshAccessToken(c.Request().Context(), token.Value)
	if err != nil {
		return err
	}

	h.setRefreshTokenCookie(c, refreshToken)
	h.setAccessTokenCookie(c, accessToken)

	return response.OK(c, response.Success(&response.RefreshRes{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}))
}

// RegisterAuthRoutes registers all auth-related routes.
func RegisterAuthRoutes(r *echo.Group, h AuthHandler) {
	group := r.Group("/auth")
	group.POST("/captcha", h.SendCaptcha)
	group.POST("/register", h.Register)
	group.POST("/login", h.Login)
	group.POST("/logout", h.Logout)
	group.POST("/refresh", h.Refresh)
}

// toLoginRes converts an AuthResult into a LoginRes.
func toLoginRes(result *service.AuthResult) *response.LoginRes {
	return &response.LoginRes{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		UUID:         result.UUID,
		Avatar:       result.Avatar,
		Username:     result.Username,
		Role:         result.Role,
	}
}

func (h *authHandler) setTokenCookie(c *echo.Context, name string, value string, exp time.Duration) {
	c.SetCookie(&http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(exp),
		HttpOnly: true,
		Secure:   h.cfg.App.IsProd(),
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *authHandler) clearTokenCookie(c *echo.Context, name string) {
	c.SetCookie(&http.Cookie{
		Name:     name,
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.App.IsProd(),
		SameSite: http.SameSiteStrictMode,
	})
}

// setRefreshTokenCookie sets the refresh token as an HTTP-only cookie.
func (h *authHandler) setRefreshTokenCookie(c *echo.Context, value string) {
	h.setTokenCookie(c, "refreshToken", value, h.cfg.JWT.RefreshExpiration)
}

func (h *authHandler) setAccessTokenCookie(c *echo.Context, value string) {
	h.setTokenCookie(c, "accessToken", value, h.cfg.JWT.AccessExpiration)
}
