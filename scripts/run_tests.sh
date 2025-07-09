#!/bin/bash

set -e

echo "🧪 Запуск всех тестов системы комментариев..."

cd "$(dirname "$0")/.."

echo "📦 Проверка зависимостей..."
go mod tidy

echo "🔍 Запуск тестов сущностей..."
go test -v ./internal/domain/entities/... -count=1

echo "🔧 Запуск тестов системы ошибок..."
go test -v ./internal/errors/... -count=1

echo "⚙️ Запуск тестов сервисов..."
go test -v ./internal/domain/services/... -count=1

echo "🌟 Запуск интеграционных тестов..."
go test -v ./tests/... -count=1

echo "📊 Генерация отчета о покрытии..."
go test -coverprofile=coverage.out ./internal/domain/entities/... ./internal/errors/... ./internal/domain/services/... ./tests/...

echo "📈 Просмотр покрытия..."
go tool cover -html=coverage.out -o coverage.html

echo "✅ Все тесты завершены успешно!"
echo "📋 Отчет о покрытии сохранен в coverage.html"

echo ""
echo "📊 Сводка покрытия:"
go tool cover -func=coverage.out | grep total 