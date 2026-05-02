package config

import (
	"fmt"

	gwucfg "github.com/harrydayexe/GoWebUtilities/config"
)

// EnvironmentConfig reads the runtime environment from the ENVIRONMENT env var
// via gowebutilities config.ParseConfig. Only the Environment field is parsed,
// so it does not conflict with the CLI-flag-driven port/timeout settings.
type EnvironmentConfig struct {
	// Environment is read from the ENVIRONMENT env var. Valid values are
	// "local" (default), "test", and "production".
	Environment gwucfg.Environment `env:"ENVIRONMENT" envDefault:"local"`
}

// Validate ensures Environment is one of the gowebutilities-defined constants.
func (c EnvironmentConfig) Validate() error {
	switch c.Environment {
	case gwucfg.Local, gwucfg.Test, gwucfg.Production:
		return nil
	default:
		return fmt.Errorf("invalid ENVIRONMENT %q (must be local, test, or production)", c.Environment)
	}
}
