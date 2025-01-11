package config

import (
	"flag"
	"os"
)

// Config - структура для хранения конфигурации сервиса.
type Config struct {
	Address         string // Адрес запуска HTTP-сервера
	BaseURL         string // Базовый адрес результирующего сокращённого URL
	FileStoragePath string
}

// InitConfig - функция для инициализации конфигурации из аргументов командной строки.
func InitConfig() (*Config, error) {
	addressFlag := flag.String("a", "localhost:8080", "HTTP server address")
	baseURLFlag := flag.String("b", "http://localhost:8080", "Base URL for the shortened URL")
	fileStoragePathFlag := flag.String("f", "default_storage.json", "Path to the file for storing URLs")

	flag.Parse()

	var address = getValue("SERVER_ADDRESS", addressFlag)
	var baseURL = getValue("BASE_URL", baseURLFlag)
	var fileStoragePath = getValue("FILE_STORAGE_PATH", fileStoragePathFlag)

	return &Config{
		Address:         address,
		BaseURL:         baseURL,
		FileStoragePath: fileStoragePath,
	}, nil
}

func getValue(envVarKey string, flagValue *string) string {
	envVar, ok := os.LookupEnv(envVarKey)
	if ok {
		return envVar
	}

	return *flagValue
}
