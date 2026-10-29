package service

import (
	"bytes"
	"fmt"
	"html/template"

	"blog-server/internal/platform/config"
	"blog-server/internal/shared/errx"
	"blog-server/templates"

	"gopkg.in/gomail.v2"
)

// MailService defines the interface for email sending operations.
type MailService interface {
	Send(to, subject, templateName string, data any) error
}

// mailService implements the MailService interface.
type mailService struct {
	dialer   *gomail.Dialer
	template *template.Template
	from     string
}

// NewEmailService creates and returns a new MailService instance.
func NewEmailService(cfg *config.Config) (MailService, error) {
	mailCfg := cfg.Email
	dialer := gomail.NewDialer(mailCfg.Host, mailCfg.Port, mailCfg.Username, mailCfg.Password)

	t, err := template.ParseFS(templates.FS, "*.html")
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return &mailService{
		dialer:   dialer,
		template: t,
		from:     mailCfg.From,
	}, nil
}

// Send sends an email with the given template and data.
func (s *mailService) Send(to, subject, templateName string, data any) error {
	var body bytes.Buffer
	if err := s.template.ExecuteTemplate(&body, templateName, data); err != nil {
		return errx.New(errx.CodeInternalError, fmt.Errorf("failed to execute email template %s: %w", templateName, err))
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body.String())

	if err := s.dialer.DialAndSend(m); err != nil {
		return errx.New(errx.CodeExternalError, fmt.Errorf("failed to send email to %s: %w", to, err))
	}

	return nil
}
