package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DB     DBConfig
	GRPC   GRPCConfig
	Server ServerConfig
	Log    LogConfig
}

type DBConfig struct {
	Host           string
	Port           int
	User           string
	Password       string
	Name           string
	SSLMode        string
	MaxConns       int
	MinConns       int
	ConnectTimeout time.Duration
}

type GRPCConfig struct {
	Host string
	Port int
}

type ServerConfig struct {
	ShutdownTimeout time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

// отсюда все запускаем по приоритету и валидации
func Load(path string) (*Config, error) {
	fileVars, err := LoadEnvFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		DB: DBConfig{
			Host:           getString("DB_HOST", fileVars, DefaultDBHOST),
			Port:           getInt("DB_PORT", fileVars, DefaultDBPort),
			User:           getString("DB_USER", fileVars, DefaultDBUser),
			Password:       getString("DB_PASSWORD", fileVars, DefaultDBPassword),
			Name:           getString("DB_NAME", fileVars, DefaultDBName),
			SSLMode:        getString("DB_SSL_MODE", fileVars, DefaultSSLMode),
			MinConns:       getInt("DB_MIN_CONNS", fileVars, DefaultDBMinConnections),
			MaxConns:       getInt("DB_MAX_CONNS", fileVars, DefaultDBMaxConnections),
			ConnectTimeout: getDuration("DB_CONNECT_TIMEOUT", fileVars, DefaultDBConnectTimeout),
		},
		GRPC: GRPCConfig{
			Host: getString("GRPC_HOST", fileVars, DefaultGRPCHost),
			Port: getInt("GRPC_PORT", fileVars, DefaultGRPCPort),
		},
		Server: ServerConfig{
			ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", fileVars, DefaultShutdownTime),
		},
		Log: LogConfig{
			Level:  getString("LOG_LEVEL", fileVars, DefaultLogLevel),
			Format: getString("LOG_FORMAT", fileVars, DefaultLogFormat),
		},
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config bad validation: %w", err)
	}
	return cfg, nil
}

// проверяем критичные поля методом для конфига
func (c *Config) Validate() error {
	if c.DB.Host == "" {
		return errors.New("DB_HOST is required")
	}
	if c.DB.Name == "" {
		return errors.New("DB_USER is required")
	}
	if c.DB.Password == "" {
		return fmt.Errorf("DB_PASSWORD cannot be empty")
	}
	if c.DB.Port < 1 || c.DB.Port > 65535 {
		return fmt.Errorf("DB_PORT should be between 1 and 65535, got %d", c.DB.Port)
	}
	if c.DB.MaxConns < c.DB.MinConns {
		return fmt.Errorf("DB_MAX_CONNS (%d) must be >= DB_MIN_CONNS (%d)", c.DB.MaxConns, c.DB.MinConns)
	}
	if c.GRPC.Port < 1 || c.GRPC.Port > 65535 {
		return fmt.Errorf("GRPC_PORT must be in range 1-65535, got %d", c.GRPC.Port)
	}
	if c.Server.ShutdownTimeout <= 0 {
		return fmt.Errorf("SHUTDOWN_TIMEOUT must be positive, got %s", c.Server.ShutdownTimeout)
	}
	switch c.Log.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("LOG_LEVEL must be one of debug|info|warn|error, got %q", c.Log.Level)
	}

	switch c.Log.Format {
	case "json", "text":
	default:
		return fmt.Errorf("LOG_FORMAT must be one of json|text, got %q", c.Log.Format)
	}
	return nil
}

func LoadEnvFile(path string) (map[string]string, error) {
	fileVars, err := godotenv.Read(path) //написать себе про godotenv для чего нужен
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	return fileVars, nil
}

func getString(key string, fileVars map[string]string, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if v, ok := fileVars[key]; ok && v != "" {
		return v
	}
	return def
}

func getInt(key string, fileVars map[string]string, def int) int {
	raw := getString(key, fileVars, "")
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}

func getDuration(key string, fileVars map[string]string, def time.Duration) time.Duration {
	raw := getString(key, fileVars, "")
	if raw == "" {
		return def
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		return def
	}
	return v
}
