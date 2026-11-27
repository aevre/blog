package email

import (
	"context"
	"errors"
	"fmt"

	"blog-server/internal/platform/config"

	"github.com/resend/resend-go/v4"
)

type resendSender struct {
	client *resend.Client
	from   string
}

func NewResendSender(cfg config.EmailConfig) (Sender, error) {
	return &resendSender{
		client: resend.NewClient(cfg.APIKey),
		from:   cfg.From,
	}, nil
}

func (s *resendSender) Send(ctx context.Context, message Message) error {
	from := message.From
	if from == "" {
		from = s.from
	}
	if len(message.To) == 0 {
		return errors.New("email: at least one recipient is required")
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      message.To,
		Subject: message.Subject,
		Html:    message.HTML,
		Text:    message.Text,
	}

	if _, err := s.client.Emails.SendWithContext(ctx, params); err != nil {
		return fmt.Errorf("email: send message: %w", err)
	}

	return nil
}
