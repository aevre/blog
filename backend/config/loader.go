package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	globalConfig *Config
	once         sync.Once
	mu           sync.RWMutex
)

// Options defines configuration loading behavior such as file location,
// environment prefix, and hot-reload settings.
type Options struct {
	ConfigFile string
	ConfigType string
	EnvPrefix  string
	WatchFile  bool
	OnChange   func(cfg *Config)
}

// Load initializes the global configuration instance using the provided options.
//
// It is safe to call multiple times, but only the first invocation takes effect.
func Load(opts ...Options) (*Config, error) {
	var err error
	log.Println("[config] loading...")
	log.Printf("[config] opts: %#v\n", opts)
	once.Do(func() {
		globalConfig, err = load(mergeOptions(opts...))
	})
	if err != nil {
		return nil, err
	}
	return globalConfig, nil
}

// MustLoad behaves like Load but terminates the program if configuration
// loading fails.
func MustLoad(opts ...Options) *Config {
	cfg, err := Load(opts...)
	if err != nil {
		log.Fatalf("[config] load failed: %v", err)
	}
	return cfg
}

// Get returns the globally initialized configuration instance.
//
// It panics if Load has not been called before usage.
func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	if globalConfig == nil {
		panic("[config] not initialized, call Load() first")
	}
	return globalConfig
}

// load performs the actual configuration loading process using viper.
// It supports file-based config, environment variables, validation,
// and optional hot-reloading.
func load(opt Options) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(opt.ConfigFile)

	// Configure environment variable support with prefix and key mapping.
	v.SetEnvPrefix(opt.EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read configuration file if available.
	if err := v.ReadInConfig(); err != nil {
		// If config file is missing, fallback to environment variables.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		log.Println("[config] no config file found, using env vars and defaults")
	} else {
		log.Printf("[config] loaded from: %s", v.ConfigFileUsed())
	}

	// Decode configuration into strongly typed struct.
	cfg, err := decode(v)
	if err != nil {
		return nil, err
	}

	// Validate required configuration fields.
	if err := validate(cfg); err != nil {
		return nil, err
	}

	// Enable hot-reloading if requested.
	if opt.WatchFile {
		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			log.Printf("[config] file changed: %s", e.Name)
			newCfg, err := decode(v)
			if err == nil {
				err = validate(newCfg)
			}
			if err != nil {
				log.Printf("[config] reload error: %v", err)
				return
			}
			mu.Lock()
			globalConfig = newCfg
			mu.Unlock()
			if opt.OnChange != nil {
				opt.OnChange(newCfg)
			}
		})
	}

	return cfg, nil
}

// decode unmarshals viper configuration into the Config struct.
func decode(v *viper.Viper) (*Config, error) {
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}

// mergeOptions merges user-provided options with default values.
// Environment variable CONFIG_FILE can override the default config file path.
func mergeOptions(opts ...Options) Options {
	opt := Options{
		ConfigFile: "config.toml",
		ConfigType: "toml",
		EnvPrefix:  "APP",
	}
	if len(opts) == 0 {
		return opt
	}

	o := opts[0]
	if o.ConfigFile != "" {
		opt.ConfigFile = o.ConfigFile
	}
	if o.ConfigType != "" {
		opt.ConfigType = o.ConfigType
	}
	if o.EnvPrefix != "" {
		opt.EnvPrefix = o.EnvPrefix
	}
	opt.WatchFile = o.WatchFile
	opt.OnChange = o.OnChange

	if envFile := os.Getenv("CONFIG_FILE"); envFile != "" {
		opt.ConfigFile = envFile
	}

	return opt
}
