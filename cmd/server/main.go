package main

import (
	"context"
	"fmt"
	"log"

	"github.com/andranikuz/gophkeeper/internal/server"
)

func main() {
	// Инициализируем сервер.
	s, err := server.NewServer(context.Background())
	if err != nil {
		log.Fatalf("Ошибка инициализации сервера: %v", err)
	}
	fmt.Println("Запуск сервера:")
	// Запускаем HTTP-сервер.
	if err := s.Run(); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
	log.Printf("Сервер запущен")
}
