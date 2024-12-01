package config

import (
    "flag"
    "fmt"
)

// Config - структура для хранения конфигурации сервиса
type Config struct {
    Address      string // Адрес запуска HTTP-сервера
    BaseURL      string // Базовый адрес результирующего сокращённого URL
}

// InitConfig - функция для инициализации конфигурации из аргументов командной строки
func InitConfig() (*Config, error) {
    address := flag.String("a", "localhost:8080", "HTTP server address")
    baseURL := flag.String("b", "http://localhost:8080/", "Base URL for the shortened URL")

    flag.Parse()

    if *baseURL == "" {
        return nil, fmt.Errorf("base URL cannot be empty")
    }

    return &Config{
        Address: *address,
        BaseURL: *baseURL,
    }, nil
}
