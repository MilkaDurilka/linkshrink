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
	DataBaseDSN     string
}

// InitConfig - функция для инициализации конфигурации из аргументов командной строки.
func InitConfig() (*Config, error) {
	addressFlag := flag.String("a", "localhost:8080", "HTTP server address")
	baseURLFlag := flag.String("b", "http://localhost:8080", "Base URL for the shortened URL")
	fileStoragePathFlag := flag.String("f", "default_storage.json", "Path to the file for storing URLs")
	dataBaseDSNFlag := flag.String("d", "", "PostgreSQL connection string")

	flag.Parse()

	address := getValue("SERVER_ADDRESS", addressFlag)
	baseURL := getValue("BASE_URL", baseURLFlag)
	fileStoragePath := getValue("FILE_STORAGE_PATH", fileStoragePathFlag)
	dataBaseDSN := getValue("DATABASE_DSN", dataBaseDSNFlag)

	return &Config{
		Address:         address,
		BaseURL:         baseURL,
		FileStoragePath: fileStoragePath,
		DataBaseDSN:     dataBaseDSN,
	}, nil
}

func getValue(envVarKey string, flagValue *string) string {
	envVar, ok := os.LookupEnv(envVarKey)
	if ok {
		return envVar
	}

	return *flagValue
}
