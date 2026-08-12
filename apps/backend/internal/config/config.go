package config

import (
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
)

type Config struct {
	Observability *ObservabilityConfig `koanf:"observability" validate:"required"`
	Primary       Primary              `koanf:"primary" validate:"required"`
	Auth          AuthConfig           `koanf:"auth" validate:"required"`
	Redis         RedisConfig          `koanf:"redis" validate:"required"`
	Database      DatabaseConfig       `koanf:"database" validate:"required"`
	Server        ServerConfig         `koanf:"server" validate:"required"`
}

type Primary struct {
	Env string `koanf:"env" validate:"required"`
}

type ServerConfig struct {
	Port               string   `koanf:"port" validate:"required"`
	CORSAllowedOrigins []string `koanf:"cors_allowed_origins" validate:"required"`
	ReadTimeout        int      `koanf:"read_timeout" validate:"required"`
	WriteTimeout       int      `koanf:"write_timeout" validate:"required"`
	IdleTimeout        int      `koanf:"idle_timeout" validate:"required"`
}

type DatabaseConfig struct {
	Host            string `koanf:"host" validate:"required"`
	Port            string `koanf:"port" validate:"required"`
	User            string `koanf:"user" validate:"required"`
	Password        string `koanf:"password"`
	Name            string `koanf:"name" validate:"required"`
	SSLMode         string `koanf:"ssl_mode" validate:"required"`
	MaxOpenConns    int    `koanf:"max_open_conns" validate:"required"`
	MaxIdleConns    int    `koanf:"max_idle_conns" validate:"required"`
	ConnMaxLifetime int    `koanf:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime int    `koanf:"conn_max_idle_time" validate:"required"`
}

type AuthConfig struct {
	SecretKey string `koanf:"secret_key" validate:"required"`
}

type RedisConfig struct {
	Address string `koanf:"address" validate:"required"`
}

func LoadConfig() (*Config, error) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	k := koanf.New("")

	err := k.Load(env.Provider("SPNDEX_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "SPNDEX_"))
	}), nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load initial environment variables")
	}

	mainConfig := &Config{}
	err = k.Unmarshal("", mainConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to unmarshal into main config")
	}

	validate := validator.New()
	err = validate.Struct(mainConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to validate main config struct")
	}

	// Default Observability
	if mainConfig.Observability == nil {
		mainConfig.Observability = DefaultObservabilityConfig()
	}

	// Override service name and encironment from primary config
	mainConfig.Observability.Environment = mainConfig.Primary.Env
	mainConfig.Observability.ServiceName = "spndex"

	err = mainConfig.Observability.Validate()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to validate observability config")
	}

	return mainConfig, nil
}
