package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	LogLevel    string
	Environment string
}

func NewConfig() *Config {
	return &Config{
		Port:        os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		LogLevel:    os.Getenv("LOG_LEVEL"),
		Environment: os.Getenv("ENV"),
	}
}
