package config

import "os"

type Config struct {
	SlackBotToken      string
	SlackAppToken      string
	SlackSigningSecret string
	DatabasePath       string
	Port               string
}

func Load() *Config {
	return &Config{
		SlackBotToken:      getEnv("SLACK_BOT_TOKEN", ""),
		SlackAppToken:      getEnv("SLACK_APP_TOKEN", ""),
		SlackSigningSecret: getEnv("SLACK_SIGNING_SECRET", ""),
		DatabasePath:       getEnv("DATABASE_PATH", "./rotation.db"),
		Port:               getEnv("PORT", "3000"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}