.PHONY: help dev dev-up dev-down backend frontend migrate test lint clean

# Цвета для вывода
GREEN  := $(shell tput setaf 2)
YELLOW := $(shell tput setaf 3)
RESET  := $(shell tput sgr0)

help: ## Показать справку
	@echo "$(GREEN)Atlas - Корпоративный мессенджер$(RESET)"
	@echo ""
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(RESET) %s\n", $$1, $$2}'

# === Development ===

dev: dev-up backend ## Запустить всё для разработки

dev-up: ## Поднять инфраструктуру (PostgreSQL, Redis, MinIO, LiveKit)
	cd deploy && docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

dev-down: ## Остановить инфраструктуру
	cd deploy && docker-compose -f docker-compose.yml -f docker-compose.dev.yml down

dev-logs: ## Логи инфраструктуры
	cd deploy && docker-compose -f docker-compose.yml -f docker-compose.dev.yml logs -f

# === Backend ===

backend: ## Запустить backend (Go)
	cd backend && go run cmd/server/main.go

backend-build: ## Собрать backend
	cd backend && go build -o bin/server cmd/server/main.go

backend-test: ## Тесты backend
	cd backend && go test -v ./...

# === Frontend ===

frontend: ## Запустить frontend (Vue 3)
	cd frontend && npm run dev

frontend-install: ## Установить зависимости frontend
	cd frontend && npm install

frontend-build: ## Собрать frontend
	cd frontend && npm run build

# === Database ===

migrate: ## Применить миграции
	@echo "TODO: Реализовать миграции"

migrate-create: ## Создать новую миграцию (name=имя)
	@echo "TODO: Реализовать создание миграций"

# === Quality ===

lint: ## Проверить код линтерами
	cd backend && golangci-lint run
	cd frontend && npm run lint

test: backend-test ## Запустить все тесты

# === Cleanup ===

clean: ## Очистить артефакты сборки
	rm -rf backend/bin
	rm -rf frontend/dist

