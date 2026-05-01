package config_test

import (
	"testing"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	gwucfg "github.com/harrydayexe/GoWebUtilities/config"
)

// t.Setenv and t.Parallel are mutually exclusive, so these tests run serially.

func TestEnvironmentConfig_DefaultIsLocal(t *testing.T) {
	t.Setenv("ENVIRONMENT", "local")

	cfg, err := gwucfg.ParseConfig[config.EnvironmentConfig]()
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	if cfg.Environment != "local" {
		t.Errorf("default Environment = %q, want %q", cfg.Environment, "local")
	}
}

func TestEnvironmentConfig_ValidValues(t *testing.T) {
	valid := []string{"local", "test", "production"}
	for _, env := range valid {
		env := env
		t.Run(env, func(t *testing.T) {
			t.Setenv("ENVIRONMENT", env)

			cfg, err := gwucfg.ParseConfig[config.EnvironmentConfig]()
			if err != nil {
				t.Fatalf("ParseConfig(%q) error = %v", env, err)
			}
			if string(cfg.Environment) != env {
				t.Errorf("Environment = %q, want %q", cfg.Environment, env)
			}
		})
	}
}

func TestEnvironmentConfig_InvalidValueRejected(t *testing.T) {
	t.Setenv("ENVIRONMENT", "staging")

	_, err := gwucfg.ParseConfig[config.EnvironmentConfig]()
	if err == nil {
		t.Error("ParseConfig() with invalid ENVIRONMENT should return an error, got nil")
	}
}
