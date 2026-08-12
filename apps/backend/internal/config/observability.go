package config

import (
	"fmt"
	"time"
)

type ObservabilityConfig struct {
	ServiceName  string             `koanf:"service_name" validate:"required"`
	Environment  string             `koanf:"environment" validate:"required"`
	Logging      LoggingConfig      `koanf:"logging" validate:"required"`
	NewRelic     NewRelicConfig     `koanf:"new_relic" validate:"required"`
	HealthChecks HealthChecksConfig `koanf:"health_checks" validate:"required"`
}

type LoggingConfig struct {
	Level  string `koanf:"level" validate:"required"`
	Format string `koanf:"format" validate:"required"`
}

type NewRelicConfig struct {
	LicenseKey                string `koanf:"license_key" validate:"required"`
	AppLogsForwardingEnabled  bool   `koanf:"app_logs_forwarding_enabled" validate:"required"`
	DistributedTracingEnabled bool   `koanf:"distributed_tracing_enabled" validate:"required"`
	DebugLogger               bool   `koanf:"debug_logger"`
}

type HealthChecksConfig struct {
	Checks   []string      `koanf:"checks"`
	Enabled  bool          `koanf:"enabled" validate:"required"`
	Interval time.Duration `koanf:"interval" validate:"min=1s"`
	Timeout  time.Duration `koanf:"timeout" validate:"min=1s"`
}

func DefaultObservabilityConfig() *ObservabilityConfig {
	return &ObservabilityConfig{
		ServiceName: "academyos",
		Environment: "development",
		Logging: LoggingConfig{
			Level:  "debug",
			Format: "json",
		},
		NewRelic: NewRelicConfig{
			LicenseKey:                "",
			AppLogsForwardingEnabled:  true,
			DistributedTracingEnabled: true,
			DebugLogger:               false, // Defaults to false to avoid mix logs
		},
		HealthChecks: HealthChecksConfig{
			Checks:   []string{"database", "redis"},
			Enabled:  true,
			Interval: time.Second * 30,
			Timeout:  time.Second * 5,
		},
	}
}

func (c *ObservabilityConfig) Validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}

	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}

	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalod level: %s (must be one of: debug, info, warn, error)", c.Logging.Level)
	}

	return nil
}

func (c *ObservabilityConfig) GetLogLevel() string {
	switch c.Logging.Level {
	case "production":
		if c.Logging.Level == "" {
			return "info"
		}
	case "development":
		if c.Logging.Level == "" {
			return "debug"
		}
	}
	return c.Logging.Level
}

func (c *ObservabilityConfig) IsProduction() bool {
	return c.Environment == "production"
}
