package email

import (
	"blog-server/internal/platform/config"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"email",
		fx.Provide(
			NewResendSender,
			fx.Annotate(
				func(cfg *config.Config) config.EmailConfig {
					return cfg.Email
				},
			),
		),
	)
}
