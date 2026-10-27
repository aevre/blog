package config

import "testing"

func TestValidateLogConfiguration(t *testing.T) {
	valid := &Config{
		App:    AppConfig{Name: "blog", Environment: EnvDev},
		Server: ServerConfig{Port: 8080},
		Log:    LogConfig{Level: "info", Format: "console"},
	}
	if err := validate(valid); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}

	invalid := *valid
	invalid.Log = LogConfig{Level: "verbose", Format: "xml"}
	if err := validate(&invalid); err == nil {
		t.Fatal("invalid log configuration accepted")
	}
}
