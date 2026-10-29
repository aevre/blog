package config

import (
	"log"

	"go.uber.org/fx"
)

// Module registers the configuration subsystem into the Fx dependency graph.
//
// It provides a fully initialized Config instance and enables hot-reload support.
// When configuration changes are detected, it triggers a graceful application
// shutdown via fx.Shutdowner so that the process can restart with new config.
func Module() fx.Option {
	return fx.Module("config", fx.Provide(func(lc fx.Lifecycle, shutdowner fx.Shutdowner) (*Config, error) {
		return Load(Options{
			WatchFile: true,
			OnChange: func(cfg *Config) {
				log.Println("[config] changed, triggering restart")
				if err := shutdowner.Shutdown(); err != nil {
					log.Printf("[config] shutdown error: %v", err)
				}
			},
		})
	}))
}
