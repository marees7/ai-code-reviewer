package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port                 string
	Env                  string
	GithubSecret         string
	LogLevel             string
	AIProvider           string
	GithubPrivateKeyPath string
	GithubAppID          string
	GithubInstallationID string
	GitHubToken          string
	OpenAIKey            string
	OpenAIModel          string
	RedisAddr            string
	QueueType            string
	OllamaURL            string
	OllamaModel          string
}

func Load() *Config {
	return &Config{
		Port:                 getEnv("PORT", "8080"),
		Env:                  getEnv("ENV", "local"),
		GithubSecret:         getEnv("GITHUB_WEBHOOK_SECRET", ""),
		LogLevel:             getEnv("LOG_LEVEL", "debug"),
		GithubPrivateKeyPath: getEnv("GITHUB_APP_PRIVATE_KEY_PATH", ""),
		GithubAppID:          getEnv("GITHUB_APP_ID", ""),
		AIProvider:           getEnv("AI_PROVIDER", "openai"),
		OllamaURL:            getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:          getEnv("OLLAMA_MODEL", "llama3"),
		GithubInstallationID: getEnv("GITHUB_APP_INSTALLATION_ID", ""),
		OpenAIKey:            getEnv("OPENAI_KEY", ""),
		OpenAIModel:          getEnv("OPENAI_MODEL", "gpt-3.5-turbo"),
		GitHubToken:          getEnv("GITHUB_TOKEN", ""),
		RedisAddr:            getEnv("REDIS_ADDR", "localhost:6379"),
		QueueType:            getEnv("QUEUE_TYPE", "memory"), // memory | redis
	}
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("invalid env %s: %v", key, err)
	}
	return i
}
