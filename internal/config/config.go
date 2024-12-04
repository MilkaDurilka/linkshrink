package config

import (
	"flag"
	"os"
)

// Config - структура для хранения конфигурации сервиса
type Config struct {
	Address string // Адрес запуска HTTP-сервера
	BaseURL string // Базовый адрес результирующего сокращённого URL
}

// InitConfig - функция для инициализации конфигурации из аргументов командной строки
func InitConfig() (*Config, error) {
	addressFlag := flag.String("a", "localhost:8080", "HTTP server address")
	baseURLFlag := flag.String("b", "http://localhost:8080", "Base URL for the shortened URL")

	flag.Parse()

	addressEnv := os.Getenv("SERVER_ADDRESS")
	baseURLEnv := os.Getenv("BASE_URL")

	address := getValue(addressEnv, addressFlag)
	baseURL := getValue(baseURLEnv, baseURLFlag)

	return &Config{
		Address: address,
		BaseURL: baseURL,
	}, nil
}

func getValue(envVar string, flagValue *string) string {
	if envVar != "" {
		return envVar
	}

	return *flagValue
}
