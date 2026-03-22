package config

import (
	"os"
)

type Config struct {
	Port              string
	DatabaseURL       string
	RedisURL          string
	SlackSigningSecret string
	GitHubToken      string
	GitHubWebhookSecret string
	LogLevel          string
	Environment       string
}

func NewConfig() *Config {
	return &Config{
		Port:                os.Getenv("PORT"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		RedisURL:           os.Getenv("REDIS_URL"),
		SlackSigningSecret: os.Getenv("SLACK_SIGNING_SECRET"),
		GitHubToken:        os.Getenv("GITHUB_TOKEN"),
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
		LogLevel:           os.Getenv("LOG_LEVEL"),
		Environment:        os.Getenv("ENV"),
	}
}
