package config

import (
	"flag"
	"fmt"
)

// Config содержит настройки приложения.
type Config struct {
	// Server settings
	Host     string // Адрес сервера
	Port     int    // Порт сервера
	GrpcPort int    // Порт grpc сервера

	// Storage settings
	DBPath string // Путь к файлу базы данных

	// Application mode: "server" или "client"
	Mode string

	// Token settings для аутентификации (например, JWT)
	TokenSecret     string // Секрет для генерации токенов
	TokenExpiration int    // Время жизни токена в секундах
}

// LoadConfig парсит аргументы командной строки и возвращает указатель на Config.
func LoadConfig() *Config {
	var cfg Config

	flag.StringVar(&cfg.Host, "host", "127.0.0.1", "Серверный адрес")
	flag.IntVar(&cfg.Port, "port", 8080, "Серверный порт")
	flag.IntVar(&cfg.GrpcPort, "grpc-port", 50051, "Порт grpc сервера")
	flag.StringVar(&cfg.DBPath, "db", "./data/gophkeeper.db", "Путь к файлу базы данных")
	flag.StringVar(&cfg.Mode, "mode", "server", "Режим работы приложения: server или client")
	flag.StringVar(&cfg.TokenSecret, "secret", "mysecret", "Секретный ключ для генерации токенов")
	flag.IntVar(&cfg.TokenExpiration, "token-exp", 3600, "Время жизни токена (в секундах)")

	flag.Parse()

	return &cfg
}

// String возвращает строковое представление конфигурации.
func (cfg *Config) String() string {
	return fmt.Sprintf("Host: %s, Port: %d, DBPath: %s, Mode: %s, TokenSecret: %s, TokenExpiration: %d",
		cfg.Host, cfg.Port, cfg.DBPath, cfg.Mode, cfg.TokenSecret, cfg.TokenExpiration)
}
