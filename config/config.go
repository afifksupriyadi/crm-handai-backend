package config

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// DatabaseConfig holds database-related configuration.
type DatabaseConfig struct {
	DatabaseURI             string        `env:"DATABASE_URI" env-required:"true"`
	DatabaseTimeout         int           `env:"DATABASE_TIMEOUT" env-default:"60"`
	DatabaseMaxOpenConns    int           `env:"DATABASE_MAX_OPEN_CONNS" env-default:"30"`
	DatabaseMaxIdleConns    int           `env:"DATABASE_MAX_IDLE_CONNS" env-default:"10"`
	DatabaseConnMaxLifetime time.Duration `env:"DATABASE_CONN_MAX_LIFETIME" env-default:"3600s"`
}

type JWTConfig struct {
	JWTIssuer        string        `env:"JWT_ISSUER" env-default:"crm-portal"`
	JWTExpiration    time.Duration `env:"JWT_EXPIRATION" env-default:"24h"`
	JWTSecretKey     string        `env:"JWT_SECRET_KEY" env-required:"true"`
	JWTSignAlgorithm string        `env:"JWT_SIGN_ALGORITHM" env-default:"HS256"`
	JWTVersion       int           `env:"JWT_VERSION" env-default:"1"`
}

type EncryptionConfig struct {
	AESEncryptionKey string `env:"AES_ENCRYPTION_KEY" env-required:"true"`
}

type CORSConfig struct {
	AllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" env-default:"*"`
}

type LoggerConfig struct {
	Level          string `env:"LOG_LEVEL" env-default:"info"`
	Format         string `env:"LOG_FORMAT" env-default:"json"` // json or console
	AccessLog      bool   `env:"LOG_ACCESS_LOG" env-default:"true"`
	LogHeaders     bool   `env:"LOG_HEADERS" env-default:"false"`
	LogBody        bool   `env:"LOG_BODY" env-default:"false"`
	LogQueryParams bool   `env:"LOG_QUERY_PARAMS" env-default:"false"`
}

type Config struct {
	ServiceName string `env:"SERVICE_NAME" env-default:"crm-backend"`
	Env         string `env:"ENV" env-default:"dev"`
	Host        string `env:"HOST" env-default:"localhost"`
	Port        int    `env:"PORT" env-default:"8080"`
	BasePath    string `env:"BASE_PATH" env-default:"/api"`
	PublishURL  string `env:"PUBLISH_URL"`

	DefaultUserPasswordHash string        `env:"DEFAULT_USER_PASSWORD_HASH" env-required:"true"`
	ShutdownTimeout         time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"5s"`
	Timezone                string        `env:"TIMEZONE" env-default:"Asia/Jakarta"`

	DB         DatabaseConfig
	JWT        JWTConfig
	Encryption EncryptionConfig
	CORS       CORSConfig
	Logger     LoggerConfig
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &cfg, nil
}

var (
	conf *Config
	once sync.Once
)

var allowedEnvironments = []string{"dev", "prod"}

// Get returns the singleton instance of Config.
// It loads configuration from .env file first, then falls back to environment variables.
func Get() *Config {
	if conf != nil {
		return conf
	}

	var err error
	once.Do(func() {
		conf = &Config{}

		if err = cleanenv.ReadConfig(".env", conf); err != nil {
			if err = cleanenv.ReadEnv(conf); err != nil {
				panic(fmt.Sprintf("invalid configuration: %v", err))
			}
		}

		if err = conf.validateConfig(); err != nil {
			panic(fmt.Sprintf("invalid configuration: %v", err))
		}
	})

	return conf
}

// validateConfig performs validation on the configuration.
func (c *Config) validateConfig() error {
	env := strings.ToLower(c.Env)
	validEnv := false
	for _, allowed := range allowedEnvironments {
		if strings.HasPrefix(env, allowed) {
			validEnv = true
			break
		}
	}
	if !validEnv {
		return fmt.Errorf("invalid environment: %s, must be one of %v", c.Env, allowedEnvironments)
	}

	if c.Port <= 0 {
		return errors.New("port must be positive")
	}

	if c.ShutdownTimeout <= 0 {
		return errors.New("shutdown timeout must be positive")
	}

	if c.DB.DatabaseTimeout <= 0 {
		return errors.New("database timeout must be positive")
	}

	return nil
}

// IsDevelopment returns true if the environment is dev.
func (c *Config) IsDevelopment() bool {
	return strings.HasPrefix(strings.ToLower(c.Env), "dev")
}

// IsProduction returns true if the environment is production.
func (c *Config) IsProduction() bool {
	return strings.HasPrefix(strings.ToLower(c.Env), "prod")
}
