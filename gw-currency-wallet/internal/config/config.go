package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTP      HTTPConfig
	DB        DBConfig
	Exchanger ExchangerConfig
	JWT       JWTConfig
	Redis     RedisConfig
	Kafka     KafkaConfig
	Server    ServerConfig
	Log       LogConfig
}

type HTTPConfig struct {
	Host string
	Port int
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

type ExchangerConfig struct {
	Addr    string //Host:port для gRPC клиента
	Timeout time.Duration
}

type JWTConfig struct {
	Secret string
	TTL    time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	RatesTTL time.Duration
}

type KafkaConfig struct {
	Brokers []string // Парсим из строки
	Topic   string
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
		HTTP: HTTPConfig{
			Host: getString("HTTP_HOST", fileVars, DefaultHTTPHost),
			Port: getInt("HTTP_PORT", fileVars, DefaultHTTPPort),
		},
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
		Exchanger: ExchangerConfig{
			Addr:    getString("EXCHANGER_ADDR", fileVars, DefaultExchangerAddress),
			Timeout: getDuration("EXCHANGER_TIMEOUT", fileVars, DefaultExchangerTimeout),
		},
		JWT: JWTConfig{
			Secret: getString("JWT_SECRET", fileVars, DefaultJWTSecret),
			TTL:    getDuration("JWT_TTL", fileVars, DefaultJWTTTL),
		},
		Redis: RedisConfig{
			Host:     getString("REDIS_HOST", fileVars, DefaultRedisHost),
			Port:     getInt("REDIS_PORT", fileVars, DefaultRedisPort),
			Password: getString("REDIS_PASSWORD", fileVars, DefaultRedisPassword),
			DB:       getInt("REDIS_DB", fileVars, DefaultRedisDB),
			RatesTTL: getDuration("REDIS_TTL", fileVars, DefaultRedisTTL),
		},
		Kafka: KafkaConfig{
			Brokers: getStringSlice("KAFKA_BROKERS", fileVars, DefaultKafkaBrokers),
			Topic:   getString("KAFKA_TOPIC", fileVars, DefaultKafkaTopic),
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
		return errors.New("DB_PASSWORD cannot be empty")
	}
	if c.DB.Port < 1 || c.DB.Port > 65535 {
		return fmt.Errorf("DB_PORT should be between 1 and 65535, got %d", c.DB.Port)
	}
	if c.DB.MaxConns < c.DB.MinConns {
		return fmt.Errorf("DB_MAX_CONNS (%d) must be >= DB_MIN_CONNS (%d)", c.DB.MaxConns, c.DB.MinConns)
	}
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		return fmt.Errorf("HTTP_PORT should be between 1 and 65535, got %d", c.HTTP.Port)
	}
	if c.Exchanger.Addr == "" {
		return errors.New("Exchanger_ADDR is required")
	}
	if len(c.JWT.Secret) < 32 {
		return errors.New("JWT_SECRET should be at least 32 characters")
	}
	if c.JWT.TTL <= 0 {
		return errors.New("JWT_TTL must be greater than zero")
	}
	if c.Redis.RatesTTL <= 0 {
		return errors.New("REDIS_TTL must be greater than zero")
	}
	if c.Redis.Host == "" {
		return errors.New("REDIS_HOST is required")
	}
	if c.Redis.Port < 1 || c.Redis.Port > 65535 {
		return fmt.Errorf("REDIS_PORT must be 1-65535, got %d", c.Redis.Port)
	}
	if len(c.Kafka.Brokers) == 0 {
		return errors.New("KAFKA_BROKERS is required")
	}
	if c.Kafka.Topic == "" {
		return errors.New("KAFKA_TOPIC is required")
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
	fileVars, err := godotenv.Read(path)
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

func getStringSlice(key string, fileVars map[string]string, def string) []string {
	raw := getString(key, fileVars, def)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
