package config

import (
	"os"
	"strings"
)

type Config struct {
	Port                string
	DatabaseURL         string
	RedisURL            string
	KafkaBrokers        []string
	SlackSigningSecret  string
	GitHubToken         string
	GitHubWebhookSecret string
	LogLevel            string
	Environment         string
}

func NewConfig() *Config {
	kafkaBrokers := []string{}
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		for _, broker := range strings.Split(brokers, ",") {
    broker = strings.TrimSpace(broker)
    if broker != "" {
        kafkaBrokers = append(kafkaBrokers, broker)
    }
}
	}

	return &Config{
		Port:                os.Getenv("PORT"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		RedisURL:            os.Getenv("REDIS_URL"),
		KafkaBrokers:        kafkaBrokers,
		SlackSigningSecret:  os.Getenv("SLACK_SIGNING_SECRET"),
		GitHubToken:         os.Getenv("GITHUB_TOKEN"),
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
		LogLevel:            os.Getenv("LOG_LEVEL"),
		Environment:         os.Getenv("ENV"),
	}
}