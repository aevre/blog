package service

import (
	"context"
	"crypto/subtle"
	"fmt"
	"slices"
	"strings"
	"time"

	"blog-server/cache"
	"blog-server/config"
	"blog-server/datastore"
	"blog-server/entity"
	"blog-server/pkg/errx"
	"blog-server/pkg/jwt"
	"blog-server/repository"
	"blog-server/utils"

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
	HasRole(ctx context.Context, id uint, roles ...entity.UserRole) (bool, error)
	RefreshAccessToken(ctx context.Context, token string) (string, string, error)
}

// authService implements the AuthService interface.
type authService struct {
	ds          *datastore.DataStore
	rc          cache.CacheClient
	cfg         *config.Config
	userRepo    repository.UserRepo
	mailService MailService
}

// NewAuthService creates and returns a new AuthService instance.
func NewAuthService(
	ds *datastore.DataStore,
	rc cache.CacheClient,
	cfg *config.Config,
	userRepo repository.UserRepo,
	mailService MailService,
) AuthService {
	return &authService{
		ds:          ds,
		rc:          rc,
		cfg:         cfg,
		userRepo:    userRepo,
		mailService: mailService,
	}
}

// Register registers a new user and generates access/refresh tokens.
func (s *authService) Register(ctx context.Context, input *RegisterInput) (*AuthResult, error) {
	// Verify captcha
	cachedCaptcha, err := s.rc.Get(ctx, fmt.Sprintf("Register:%s", input.Email))
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, fmt.Errorf("failed to get cached captcha: %w", err))
	}
	if !strings.EqualFold(strings.TrimSpace(cachedCaptcha), strings.TrimSpace(input.Captcha)) {
		return nil, errx.New(errx.CodeInvalidParam, fmt.Errorf("invalid captcha"))
	}
	_ = s.rc.Delete(ctx, fmt.Sprintf("Register:%s", input.Email))

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	user := &entity.User{
		UUID:     uuid.New(),
		Email:    input.Email,
		Username: entity.GenerateUsername(),
	}

	var created *entity.User
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

	j := jwt.New(s.cfg.JWT)
	accessToken, refreshToken, err := j.GenerateAllTokens(created.ID, created.Role)
	if err != nil {
		return nil, err
	}

	if err := s.cacheRefreshToken(ctx, created.ID, refreshToken); err != nil {
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
	if !utils.VerifyPassword(input.Password, user.Password) {
		return nil, errx.New(errx.CodeInvalidParam, fmt.Errorf("invalid credentials"))
	}

	// Fetch full profile for the response.
	profile, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}

	j := jwt.New(s.cfg.JWT)
	accessToken, refreshToken, err := j.GenerateAllTokens(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	if err := s.cacheRefreshToken(ctx, user.ID, refreshToken); err != nil {
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
	claims, err := jwt.New(s.cfg.JWT).Parse(refreshToken)
	if err != nil {
		return err
	}
	if claims.TokenType != jwt.TokenTypeRefresh {
		return errx.New(errx.CodeUnauthorized, fmt.Errorf("refresh token required"))
	}

	if err := s.rc.Delete(ctx, fmt.Sprintf("RefreshToken:%d", claims.ID)); err != nil {
		return errx.New(errx.CodeInternalError, err)
	}
	return nil
}

// RefreshAccessToken refreshes the access token using a valid refresh token.
func (s *authService) RefreshAccessToken(ctx context.Context, token string) (string, string, error) {
	j := jwt.New(s.cfg.JWT)
	claims, err := j.Parse(token)
	if err != nil {
		return "", "", err
	}
	if claims.TokenType != jwt.TokenTypeRefresh {
		return "", "", errx.New(errx.CodeUnauthorized, fmt.Errorf("refresh token required"))
	}

	userID := claims.ID
	user, err := s.userRepo.GetAuthByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	cachedRefreshToken, err := s.rc.Get(ctx, fmt.Sprintf("RefreshToken:%d", userID))
	if err != nil {
		return "", "", errx.New(errx.CodeInternalError, err)
	}

	if subtle.ConstantTimeCompare([]byte(token), []byte(cachedRefreshToken)) == 0 {
		return "", "", errx.New(errx.CodeUnauthorized, fmt.Errorf("invalid refresh token"))
	}

	accessToken, refreshToken, err := j.GenerateAllTokens(userID, user.Role)
	if err != nil {
		return "", "", err
	}

	if err := s.cacheRefreshToken(ctx, userID, refreshToken); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// HasRole checks if a user has any of the specified roles.
func (s *authService) HasRole(ctx context.Context, id uint, roles ...entity.UserRole) (bool, error) {
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

// cacheRefreshToken stores the refresh token in Redis.
func (s *authService) cacheRefreshToken(ctx context.Context, id uint, refreshToken string) error {
	key := fmt.Sprintf("RefreshToken:%d", id)
	if err := s.rc.Set(ctx, key, refreshToken, s.cfg.JWT.RefreshExpiration); err != nil {
		return errx.New(
			errx.CodeInternalError,
			fmt.Errorf("failed to cache refresh token for user %d", id),
		)
	}
	return nil
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
	return utils.RandomString(6, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
}
