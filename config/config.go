package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	CORS     CORSConfig     `yaml:"cors"`
	Logger   LoggerConfig   `yaml:"logger"`
	Upload   UploadConfig   `yaml:"upload"`
}

type AppConfig struct {
	Name string `env:"APP_NAME" env-default:"CRM Handai"`
	Env  string `env:"APP_ENV" env-default:"development"`
	Port int    `env:"APP_PORT" env-default:"8080"`
	Host string `env:"APP_HOST" env-default:"localhost"`
}

type DatabaseConfig struct {
	Host         string        `env:"DB_HOST" env-default:"localhost"`
	Port         int           `env:"DB_PORT" env-default:"5432"`
	User         string        `env:"DB_USER" env-default:"postgres"`
	Password     string        `env:"DB_PASSWORD" env-default:"postgres"`
	Name         string        `env:"DB_NAME" env-default:"crm_handai"`
	SSLMode      string        `env:"DB_SSL_MODE" env-default:"disable"`
	MaxOpenConns int           `env:"DB_MAX_OPEN_CONNS" env-default:"25"`
	MaxIdleConns int           `env:"DB_MAX_IDLE_CONNS" env-default:"5"`
	MaxLifetime  time.Duration `env:"DB_MAX_LIFETIME" env-default:"5m"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST" env-default:"localhost"`
	Port     int    `env:"REDIS_PORT" env-default:"6379"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" env-default:"0"`
}

type JWTConfig struct {
	Secret        string        `env:"JWT_SECRET" env-required:"true"`
	AccessExpiry  time.Duration `env:"JWT_ACCESS_EXPIRY" env-default:"1h"`
	RefreshExpiry time.Duration `env:"JWT_REFRESH_EXPIRY" env-default:"168h"`
}

type CORSConfig struct {
	AllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" env-separator:","`
	AllowedMethods []string `env:"CORS_ALLOWED_METHODS" env-separator:","`
	AllowedHeaders []string `env:"CORS_ALLOWED_HEADERS" env-separator:","`
}

type LoggerConfig struct {
	Level  string `env:"LOG_LEVEL" env-default:"debug"`
	Format string `env:"LOG_FORMAT" env-default:"json"`
}

type UploadConfig struct {
	MaxSize      int64    `env:"MAX_UPLOAD_SIZE" env-default:"10485760"` // 10MB
	AllowedTypes []string `env:"ALLOWED_FILE_TYPES" env-separator:","`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &cfg, nil
}

// GetDSN returns PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

// GetRedisAddr returns Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
