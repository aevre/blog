package config

import (
	"fmt"
	"strings"
)

// validate performs basic validation on the loaded configuration.
//
// It ensures required fields are present and enforces environment-specific rules.
// In production mode, it additionally checks for sensitive configuration values
// such as database credentials and JWT secrets.
func validate(cfg *Config) error {
	var errs []string

	if cfg.App.Name == "" {
		errs = append(errs, "app.name is required")
	}
	if cfg.App.Environment != EnvDev && cfg.App.Environment != EnvProd {
		errs = append(errs, "app.environment must be development or production")
	}
	if cfg.Log.Format != "" && cfg.Log.Format != "console" && cfg.Log.Format != "json" {
		errs = append(errs, "log.format must be console or json")
	}
	switch strings.ToLower(cfg.Log.Level) {
	case "debug", "info", "warn", "error":
	default:
		errs = append(errs, "log.level must be debug, info, warn, or error")
	}
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		errs = append(errs, "server.port must be between 1 and 65535")
	}
	if cfg.App.IsProd() && cfg.JWT.Secret == "" {
		errs = append(errs, "jwt.secret is required in production")
	}
	if cfg.App.IsProd() && cfg.Database.Password == "" {
		errs = append(errs, "database.password is required in production")
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}
