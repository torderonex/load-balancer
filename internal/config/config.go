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
	HttpServer  HttpServer  `yaml:"http_server"`
	HealthCheck HealthCheck `yaml:"health_check"`
	Balancer    Balancer    `yaml:"balancer"`
}

type Balancer struct {
	Backends []string `yaml:"backends"`
	Strategy string   `yaml:"strategy"`
}

type HttpServer struct {
	Port        int           `yaml:"port" env:"HTTP_SERVER_PORT" env-default:"8080"`
	ReadTimeout time.Duration `yaml:"read_timeout" env:"HTTP_SERVER_READ_TIMEOUT" env-default:"10s"`
}

type HealthCheck struct {
	Enabled       bool `yaml:"enabled" env:"HEALTH_CHECK_ENABLED" env-default:"true"`
	CheckInterval int  `yaml:"check_interval" env:"HEALTH_CHECK_CHECK_INTERVAL" env-default:"10"`
}

type RateLimiter struct {
	Enabled bool `yaml:"enabled" env:"RATE_LIMITER_ENABLED" env-default:"true"`
	//tokens per second
	DefaultRate int `yaml:"default_rate" env:"RATE_LIMITER_DEFAULT_RATE" env-default:"100"`
	//maximum tokens per server
	Capacity int `yaml:"capacity" env:"RATE_LIMITER_CAPACITY" env-default:"100"`
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
