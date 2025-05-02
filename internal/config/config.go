package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Logger      Logger      `yaml:"logger"`
	RateLimiter RateLimiter `yaml:"rate_limit"`
	ProxyServer ProxyServer `yaml:"proxy_server"`
	ApiServer   ApiServer   `yaml:"api_server"`
	HealthCheck HealthCheck `yaml:"health_check"`
	Balancer    Balancer    `yaml:"balancer"`
	Storage     Storage     `yaml:"storage"`
}

type Storage struct {
	Type  string `yaml:"type" env:"STORAGE_TYPE" env-default:"memory"`
	Redis Redis  `yaml:"redis"`
}

type Redis struct {
	Host     string `yaml:"host" env:"REDIS_HOST" env-default:"redis"`
	Port     int    `yaml:"port" env:"REDIS_PORT" env-default:"6379"`
	DB       int    `yaml:"db" env:"REDIS_DB" env-default:"0"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	Prefix   string `yaml:"prefix" env:"REDIS_PREFIX" env-default:"lb"`
}

type Balancer struct {
	Backends []string `yaml:"backends"`
	Strategy string   `yaml:"strategy"`
}

type ProxyServer struct {
	Port        int           `yaml:"port" env:"PROXY_SERVER_PORT" env-default:"8080"`
	ReadTimeout time.Duration `yaml:"read_timeout" env:"PROXY_SERVER_READ_TIMEOUT" env-default:"10s"`
}

type ApiServer struct {
	Port        int           `yaml:"port" env:"API_SERVER_PORT" env-default:"8081"`
	ReadTimeout time.Duration `yaml:"read_timeout" env:"API_SERVER_READ_TIMEOUT" env-default:"10s"`
}

type HealthCheck struct {
	Enabled       bool          `yaml:"enabled" env:"HEALTH_CHECK_ENABLED" env-default:"true"`
	CheckInterval time.Duration `yaml:"check_interval" env:"HEALTH_CHECK_CHECK_INTERVAL" env-default:"10s"`
}

type RateLimiter struct {
	Enabled bool `yaml:"enabled" env:"RATE_LIMITER_ENABLED" env-default:"true"`
	//tokens per second
	DefaultRate int `yaml:"default_rate" env:"RATE_LIMITER_DEFAULT_RATE" env-default:"100"`
	//maximum tokens per server
	Capacity       int           `yaml:"capacity" env:"RATE_LIMITER_CAPACITY" env-default:"100"`
	RefillInterval time.Duration `yaml:"refill_interval" env:"RATE_LIMITER_REFILL_INTERVAL" env-default:"10s"`
}

type Logger struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
}

// Load загружает конфигурацию приложения.
// Порядок приоритета:
// 1. Переменные окружения (самый высокий приоритет).
// 2. Значения из YAML файла (если найден).
// 3. Значения по умолчанию (env-default).
// Функция паникует, если не удается прочитать обязательные переменные окружения (env-required).
func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("INFO: .env file not found or error loading it: %v. Relying on existing environment variables.", err)
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yml"
	}

	var cfg Config

	if _, err := os.Stat(configPath); err == nil {
		err := cleanenv.ReadConfig(configPath, &cfg)
		if err != nil {
			log.Printf("WARN: Error reading config file '%s': %v. Relying solely on environment variables.", configPath, err)
		} else {
			log.Printf("INFO: Loaded base configuration structure from file: %s", configPath)
		}
	} else if !os.IsNotExist(err) {
		log.Printf("WARN: Error accessing config file '%s': %v. Relying solely on environment variables.", configPath, err)
	} else {
		log.Printf("INFO: Configuration file not found at '%s'. Relying solely on environment variables.", configPath)
	}

	log.Printf("INFO: Reading environment variables (will override YAML values if any)...")
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("FATAL: Error reading environment variables: %v", err)
	}

	log.Printf("INFO: Configuration loaded successfully. Log Level: %s",
		cfg.Logger.Level)

	return &cfg
}
