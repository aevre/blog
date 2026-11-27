package service

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"blog-server/internal/platform/email"
	"blog-server/internal/shared/errx"
)

// EmailService defines the interface for email sending operations.
type EmailService interface {
	SendCode(ctx context.Context, address string, code string, captchaType CaptchaType) error
}

// mailService implements the MailService interface.
type mailService struct {
	email    email.Sender
	template *template.Template
}

// NewEmailService creates and returns a new MailService instance.
func NewEmailService(email email.Sender, tmpl *template.Template) EmailService {
	return &mailService{
		email,
		tmpl,
	}
}

// SendCode sends an email with the given template and data.
func (s *mailService) SendCode(ctx context.Context, address string, code string, captchaType CaptchaType) error {
	var body bytes.Buffer

	data, err := getCaptchaEmailMeta(captchaType)
	if err != nil {
		return err
	}

	data.Captcha = code

	if err := s.template.ExecuteTemplate(&body, "captcha.html", data); err != nil {
		return errx.New(errx.CodeInternalError, err)
	}

	err = s.email.Send(ctx, email.Message{
		To:      []string{address},
		Subject: data.Subject,
		HTML:    body.String(),
	})
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
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
