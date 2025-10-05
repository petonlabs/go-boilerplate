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
	Primary       Primary              `koanf:"primary" validate:"required"`
	Server        ServerConfig         `koanf:"server" validate:"required"`
	Database      DatabaseConfig       `koanf:"database" validate:"required"`
	Auth          AuthConfig           `koanf:"auth" validate:"required"`
	Redis         RedisConfig          `koanf:"redis" validate:"required"`
	Integration   IntegrationConfig    `koanf:"integration" validate:"required"`
	Observability *ObservabilityConfig `koanf:"observability"`
}

type Primary struct {
	Env string `koanf:"env" validate:"required"`
}

type ServerConfig struct {
	Port               string   `koanf:"port" validate:"required"`
	ReadTimeout        int      `koanf:"read_timeout" validate:"required"`
	WriteTimeout       int      `koanf:"write_timeout" validate:"required"`
	IdleTimeout        int      `koanf:"idle_timeout" validate:"required"`
	CORSAllowedOrigins []string `koanf:"cors_allowed_origins" validate:"required"`
}

type DatabaseConfig struct {
	Host            string `koanf:"host" validate:"required"`
	Port            int    `koanf:"port" validate:"required"`
	User            string `koanf:"user" validate:"required"`
	Password        string `koanf:"password"`
	Name            string `koanf:"name" validate:"required"`
	SSLMode         string `koanf:"ssl_mode" validate:"required"`
	MaxOpenConns    int    `koanf:"max_open_conns" validate:"required"`
	MaxIdleConns    int    `koanf:"max_idle_conns" validate:"required"`
	ConnMaxLifetime int    `koanf:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime int    `koanf:"conn_max_idle_time" validate:"required"`
}
type RedisConfig struct {
	Address string `koanf:"address" validate:"required"`
}

type IntegrationConfig struct {
	ResendAPIKey string `koanf:"resend_api_key" validate:"required"`
}

type AuthConfig struct {
	SecretKey string `koanf:"secret_key" validate:"required"`
	// PasswordResetTTL is the default TTL (in seconds) for password reset tokens
	PasswordResetTTL int `koanf:"password_reset_ttl"`
	// DeletionDefaultTTL is the default TTL (in seconds) for scheduled deletions
	DeletionDefaultTTL int `koanf:"deletion_default_ttl"`
	// WebhookSigningSecret is the Svix/Clerk signing secret used to verify incoming webhooks
	WebhookSigningSecret string `koanf:"webhook_signing_secret"`
	// WebhookToleranceSec is the allowed clock skew in seconds for webhook timestamps
	WebhookToleranceSec int `koanf:"webhook_tolerance_sec"`
	// TokenHMACSecret is the secret used to HMAC password reset tokens before storing them.
	// If empty, Auth.SecretKey will be used as a fallback.
	TokenHMACSecret string `koanf:"token_hmac_secret"`
	// AdminToken is a simple shared secret used to protect lightweight admin endpoints
	// (used only for internal tooling/tests). For production, use a stronger auth
	// mechanism or centralized secret management.
	AdminToken string `koanf:"admin_token"`
}

func LoadConfig() (*Config, error) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	k := koanf.New(".")

	// Use strings.ToLower directly instead of wrapping in lambda
	// Map environment variables to koanf keys. We want SERVER_READ_TIMEOUT
	// -> server.read_timeout so we replace the FIRST underscore with a dot
	// and lowercase the rest. The env.Provider transform runs before
	// splitting by the delimiter, so we set the delimiter to '.' and use a
	// transform that lowercases and converts the first '_' to '.'.
	// Use double underscore as a delimiter so environment variables like
	// OBSERVABILITY__NEW_RELIC__LICENSE_KEY become
	// observability.new_relic.license_key which matches the koanf struct
	// tags. Keep transform simple (lowercase) because the delimiter handles
	// splitting into segments.
	err := k.Load(env.Provider("", "__", strings.ToLower), nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not load initial env variables")
	}

	mainConfig := &Config{}

	err = k.Unmarshal("", mainConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not unmarshal main config")
	}

	validate := validator.New()

	err = validate.Struct(mainConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("config validation failed")
	}

	// Set default observability config if not provided
	if mainConfig.Observability == nil {
		mainConfig.Observability = DefaultObservabilityConfig()
	}

	// Override service name and environment from primary config
	mainConfig.Observability.ServiceName = "boilerplate"
	mainConfig.Observability.Environment = mainConfig.Primary.Env

	if err := mainConfig.Observability.Validate(); err != nil {
		logger.Fatal().Err(err).Msg("invalid observability config")
	}

	return mainConfig, nil
}
