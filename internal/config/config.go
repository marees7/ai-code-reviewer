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
	GithubPrivateKeyPath string
	GithubAppID          string
	GithubInstallationID string
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		Env:          getEnv("ENV", "local"),
		GithubSecret: getEnv("GITHUB_WEBHOOK_SECRET", ""),
		LogLevel:     getEnv("LOG_LEVEL", "debug"),
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
