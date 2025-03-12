#!/bin/bash
# Скрипт сборки проекта GophKeeper
# Выполняет кросс-компиляцию для Linux, macOS (darwin) и Windows
# Бинарные файлы сохраняются в директорию ../build

set -e

# Определяем директорию для сборки
OUTPUT_DIR="./build"
mkdir -p "$OUTPUT_DIR"

echo "Сборка серверной и клиентской частей для Linux..."
GOOS=linux GOARCH=amd64 go build -o "$OUTPUT_DIR/gophkeeper-server-linux" ./cmd/server
GOOS=linux GOARCH=amd64 go build -o "$OUTPUT_DIR/gophkeeper-client-linux" -ldflags "-X 'main.Version=1.0.0' -X 'main.BuildDate=$(date +%Y-%m-%d)'" ./cmd/client

echo "Сборка серверной и клиентской частей для macOS..."
GOOS=darwin GOARCH=amd64 go build -o "$OUTPUT_DIR/gophkeeper-server-darwin" ./cmd/server
GOOS=darwin GOARCH=amd64 go build -o "$OUTPUT_DIR/gophkeeper-client-darwin" -ldflags "-X 'main.Version=1.0.0' -X 'main.BuildDate=$(date +%Y-%m-%d)'" ./cmd/client

echo "Сборка серверной и клиентской частей для Windows..."
GOOS=windows GOARCH=amd64 go build -o "$OUTPUT_DIR/gophkeeper-server-windows.exe" ./cmd/server
GOOS=windows GOARCH=amd64 go build -o "$OUTPUT_DIR/gophkeeper-client-windows.exe" -ldflags "-X 'main.Version=1.0.0' -X 'main.BuildDate=$(date +%Y-%m-%d)'" ./cmd/client

echo "Сборка завершена. Бинарные файлы находятся в директории $OUTPUT_DIR"
