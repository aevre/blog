// Package service implements account and authentication use cases.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	accountmodel "blog-server/internal/account/model"
	"blog-server/internal/account/repository"
	"blog-server/internal/platform/cache"
	"blog-server/internal/platform/config"
	"blog-server/internal/platform/database"
	"blog-server/internal/shared/errx"
	"blog-server/internal/shared/jwt"
	"blog-server/internal/shared/randomx"

	"github.com/google/uuid"
)

// CaptchaType represents different types of captcha operations.
type CaptchaType string

const (
	// Register is for user registration captcha.
	Register CaptchaType = "Register"
	// PasswordReset is for password reset captcha.
	PasswordReset CaptchaType = "PasswordReset"
	// ChangeEmail is for email change captcha.
	ChangeEmail CaptchaType = "ChangeEmail"
)

// AuthMailData represents the data required to send a captcha email.
type AuthMailData struct {
	Title   string
	Type    string
	Content string
	Subject string
	Captcha string
}

// RegisterInput groups all parameters for user registration.
type RegisterInput struct {
	Email    string
	Password string
	Captcha  string
}

// LoginInput groups all parameters for user login.
type LoginInput struct {
	Email    string
	Password string
}

// AuthResult is the unified return type for Register / Login.
// It carries both tokens and the user profile so the handler can
// populate the full LoginRes without a second query.
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	UUID         string
	Avatar       *string
	Username     string
	Role         string
}

// AuthService defines the interface for authentication services.
type AuthService interface {
	SendCaptchaMail(ctx context.Context, to string, captchaType CaptchaType) error
	Register(ctx context.Context, input *RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, input *LoginInput) (*AuthResult, error)
	Logout(ctx context.Context, refreshToken string) error
	HasRole(ctx context.Context, id uint, roles ...accountmodel.UserRole) (bool, error)
	RefreshAccessToken(ctx context.Context, token string) (string, string, error)
}

type sessionRepo interface {
	Create(context.Context, uuid.UUID, uint, string, time.Time) error
	Rotate(context.Context, string, string, time.Time) (uint, uuid.UUID, bool, error)
	Revoke(context.Context, string) error
}

// authService implements the AuthService interface.
type authService struct {
	ds          *database.DataStore
	rc          cache.Store
	sessions    sessionRepo
	cfg         *config.Config
	userRepo    repository.UserRepo
	mailService MailService
}

// NewAuthService creates and returns a new AuthService instance.
func NewAuthService(
	ds *database.DataStore,
	rc *cache.Client,
	cfg *config.Config,
	userRepo repository.UserRepo,
	mailService MailService,
	sessions *repository.SessionRepo,
) AuthService {
	return &authService{
		ds:          ds,
		rc:          rc,
		cfg:         cfg,
		userRepo:    userRepo,
		mailService: mailService,
		sessions:    sessions,
	}
}

// Register registers a new user and generates access/refresh tokens.
func (s *authService) Register(ctx context.Context, input *RegisterInput) (*AuthResult, error) {
	// Verify captcha
	cachedCaptcha, err := s.rc.Get(ctx, fmt.Sprintf("Register:%s", input.Email))
	if errors.Is(err, cache.ErrMiss) {
		return nil, errx.New(errx.CodeInvalidParam, fmt.Errorf("invalid captcha"))
	}
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, fmt.Errorf("failed to get cached captcha: %w", err))
	}
	if !strings.EqualFold(strings.TrimSpace(cachedCaptcha), strings.TrimSpace(input.Captcha)) {
		return nil, errx.New(errx.CodeInvalidParam, fmt.Errorf("invalid captcha"))
	}
	_ = s.rc.Delete(ctx, fmt.Sprintf("Register:%s", input.Email))

	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	user := &accountmodel.User{
		UUID:     uuid.New(),
		Email:    input.Email,
		Username: accountmodel.GenerateUsername(),
	}

	var created *accountmodel.User
	err = s.ds.WithTx(ctx, func(ctx context.Context) error {
		var err error
		created, err = s.userRepo.Create(ctx, user, hashedPassword)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	sessionID := uuid.New()

	j := jwt.New(s.cfg.JWT)
	accessToken, refreshToken, err := generateSessionTokens(j, created.ID, created.Role, sessionID)
	if err != nil {
		return nil, err
	}

	if err := s.createSession(ctx, created.ID, sessionID, refreshToken); err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UUID:         created.UUID.String(),
		Avatar:       created.Avatar,
		Username:     created.Username,
		Role:         string(created.Role),
	}, nil
}

// Login logs in a user and generates access/refresh tokens.
func (s *authService) Login(ctx context.Context, input *LoginInput) (*AuthResult, error) {
	user, err := s.userRepo.GetAuthByEmail(ctx, input.Email)
	if err != nil {
		if errx.ToAppError(err).Code == errx.CodeNotFound {
			return nil, errx.New(errx.CodeInvalidParam, fmt.Errorf("invalid credentials"))
		}
		return nil, err
	}
	if !verifyPassword(input.Password, user.Password) {
		return nil, errx.New(errx.CodeInvalidParam, fmt.Errorf("invalid credentials"))
	}

	// Fetch full profile for the response.
	profile, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}

	sessionID := uuid.New()

	j := jwt.New(s.cfg.JWT)
	accessToken, refreshToken, err := generateSessionTokens(j, user.ID, user.Role, sessionID)
	if err != nil {
		return nil, err
	}

	if err := s.createSession(ctx, user.ID, sessionID, refreshToken); err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UUID:         profile.UUID.String(),
		Avatar:       profile.Avatar,
		Username:     profile.Username,
		Role:         string(profile.Role),
	}, nil
}

// SendCaptchaMail generates a captcha, stores it in Redis, and sends an email.
func (s *authService) SendCaptchaMail(ctx context.Context, to string, captchaType CaptchaType) error {
	if captchaType == "" {
		captchaType = Register
	}

	mailData, err := getCaptchaEmailMeta(captchaType)
	if err != nil {
		return err
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, to)
	if err != nil {
		return err
	}

	// Do not reveal whether the address is registered (user enumeration).
	// Reject mismatched states with the same generic error in both directions.
	shouldSend := (captchaType == Register && !exists) || (captchaType != Register && exists)
	if !shouldSend {
		return errx.New(errx.CodeInvalidParam, fmt.Errorf("captcha request rejected"))
	}

	captcha := generateCaptcha()
	mailData.Captcha = captcha

	if err := s.rc.Set(ctx, fmt.Sprintf("%s:%s", captchaType, to), captcha, 5*time.Minute); err != nil {
		return errx.New(errx.CodeInternalError, err)
	}

	templateName := "captcha.html"
	if err := s.mailService.Send(to, mailData.Subject, templateName, mailData); err != nil {
		return err
	}

	return nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	if err := s.sessions.Revoke(ctx, tokenHash(refreshToken)); err != nil {
		return errx.New(errx.CodeInternalError, err)
	}
	return nil
}

// RefreshAccessToken refreshes the access token using a valid refresh token.
func (s *authService) RefreshAccessToken(ctx context.Context, token string) (string, string, error) {
	j := jwt.New(s.cfg.JWT)
	claims, err := parseRefreshToken(j, token)
	if err != nil {
		return "", "", err
	}
	sid, err := uuid.Parse(claims.SessionID)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token session ID: %w", err)
	}

	refreshToken, err := j.GenerateRefreshToken(claims.ID, claims.Role, sid)
	if err != nil {
		return "", "", err
	}
	next, err := parseRefreshToken(j, refreshToken)
	if err != nil {
		return "", "", err
	}
	userID, sessionID, rotated, err := s.sessions.Rotate(ctx, tokenHash(token), tokenHash(refreshToken), next.ExpiresAt.Time)
	if err != nil {
		return "", "", errx.New(errx.CodeInternalError, err)
	}
	if !rotated {
		return "", "", errx.New(errx.CodeUnauthorized, fmt.Errorf("invalid refresh token"))
	}
	user, err := s.userRepo.GetAuthByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	accessToken, err := jwt.New(s.cfg.JWT).GenerateAccessToken(userID, user.Role, sessionID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// HasRole checks if a user has any of the specified roles.
func (s *authService) HasRole(ctx context.Context, id uint, roles ...accountmodel.UserRole) (bool, error) {
	user, err := s.userRepo.GetAuthByID(ctx, id)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, errx.New(errx.CodeNotFound, fmt.Errorf("user %d not found", id))
	}

	if slices.Contains(roles, user.Role) {
		return true, nil
	}
	return false, nil
}

// createSession persists only the refresh token hash.
func (s *authService) createSession(ctx context.Context, id uint, sessionID uuid.UUID, refreshToken string) error {
	claims, err := parseRefreshToken(jwt.New(s.cfg.JWT), refreshToken)
	if err != nil {
		return err
	}
	if err := s.sessions.Create(ctx, sessionID, id, tokenHash(refreshToken), claims.ExpiresAt.Time); err != nil {
		return errx.New(
			errx.CodeInternalError,
			fmt.Errorf("failed to persist refresh session for user %d: %w", id, err),
		)
	}
	return nil
}

func parseRefreshToken(j jwt.Jwt, token string) (*jwt.Claims, error) {
	claims, err := j.Parse(token)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != jwt.TokenTypeRefresh || claims.ID == 0 || claims.SessionID == "" || claims.RegisteredClaims.ID == "" {
		return nil, errx.New(errx.CodeUnauthorized, fmt.Errorf("invalid refresh token claims"))
	}
	return claims, nil
}

func generateSessionTokens(j jwt.Jwt, id uint, role accountmodel.UserRole, sessionID uuid.UUID) (string, string, error) {
	return j.GenerateAllTokens(id, role, sessionID)
}

func tokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

var captchaMetaMap = map[CaptchaType]*AuthMailData{
	Register: {
		Subject: "[Immortal's Blog] Email Verification",
		Title:   "Verify Your Registration",
		Type:    "Registration",
		Content: "Please use the following verification code to complete your registration.",
	},
	PasswordReset: {
		Subject: "[Immortal's Blog] Password Reset",
		Title:   "Verify Your Password Reset Request",
		Type:    "Password Reset",
		Content: "You are attempting to reset your password. Please use this verification code to proceed.",
	},
	ChangeEmail: {
		Subject: "[Immortal's Blog] Change Email Address",
		Title:   "Verify Your New Email Address",
		Type:    "Email Change",
		Content: "You are attempting to change your email address. Please use this verification code to verify your new email.",
	},
}

// getCaptchaEmailMeta returns email metadata based on captcha type.
func getCaptchaEmailMeta(captchaType CaptchaType) (*AuthMailData, error) {
	data, ok := captchaMetaMap[captchaType]
	if !ok {
		return nil, errx.New(
			errx.CodeInvalidParam,
			fmt.Errorf("unknown captcha type: %s", captchaType),
		)
	}
	copyData := *data
	return &copyData, nil
}

// generateCaptcha generates a random 6-character alphanumeric captcha.
func generateCaptcha() string {
	return randomx.String(6, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
}
