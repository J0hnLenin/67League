SHELL := /bin/bash
GO               ?= go
MOCKERY_VERSION  ?= v3.8.0
GOLANGCI_VERSION ?= v1.59.1
MIGRATE_VERSION  ?= v4.17.1
DB_DSN           ?= postgres://liga:liga@localhost:5432/liga67?sslmode=disable

.PHONY: help
help: ## Показать список целей
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

## --- Инструменты -------------------------------------------------------

.PHONY: tools
tools: ## Установить инструменты разработки
	$(GO) install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_VERSION)

## --- Генерация ---------------------------------------------------------

.PHONY: mocks
mocks: ## Сгенерировать моки из интерфейсов ports
	@echo ">>> generating mocks"
	$(GO) run github.com/vektra/mockery/v3@$(MOCKERY_VERSION)
	@echo ">>> mocks generated in internal/testsupport/mocks"

.PHONY: generate
generate: mocks ## Всё, что генерируется

## --- Тесты -------------------------------------------------------------

.PHONY: test
test: ## Все тесты с race detector
	$(GO) test ./... -race -count=1

.PHONY: test-usecases
test-usecases: ## Только тесты use cases, verbose
	$(GO) test ./internal/application/usecases/... -race -count=1 -v

.PHONY: test-domain
test-domain: ## Только домен, verbose
	$(GO) test ./internal/domain/... -race -count=1 -v

.PHONY: test-integration
test-integration: ## Интеграционные тесты (testcontainers)
	$(GO) test ./internal/infrastructure/... -race -count=1 -v

.PHONY: test-cover
test-cover: ## Покрытие + HTML-отчёт
	$(GO) test ./... -race -count=1 -coverprofile=coverage.out -covermode=atomic
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo ">>> отчёт: coverage.html"

## --- Качество ----------------------------------------------------------

.PHONY: fmt
fmt: ## gofmt по проекту
	$(GO) fmt ./...

.PHONY: vet
vet: ## go vet
	$(GO) vet ./...

.PHONY: lint
lint: ## golangci-lint
	golangci-lint run

.PHONY: check
check: fmt vet lint test ## Локальный чек перед PR

## --- Сборка / очистка --------------------------------------------------

.PHONY: build
build: ## Собрать все пакеты
	$(GO) build ./...

.PHONY: clean
clean: ## Убрать сгенерированное
	rm -rf internal/testsupport/mocks coverage.out coverage.html bin/

## --- База данных -------------------------------------------------------

.PHONY: db-up
db-up: ## Поднять Postgres в Docker
	docker compose up -d postgres

.PHONY: db-down
db-down: ## Остановить Postgres
	docker compose down

.PHONY: db-logs
db-logs: ## Логи Postgres
	docker compose logs -f postgres

.PHONY: migrate-up
migrate-up: ## Применить все миграции к локальной БД
	$(GO) run github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION) \
		-path migrations \
		-database "$(DB_DSN)" up

.PHONY: migrate-down
migrate-down: ## Откатить одну миграцию
	$(GO) run github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION) \
		-path migrations \
		-database "$(DB_DSN)" down 1

.PHONY: migrate-create
migrate-create: ## Новая миграция: make migrate-create NAME=add_something
	@if [ -z "$(NAME)" ]; then echo "usage: make migrate-create NAME=add_something"; exit 1; fi
	$(GO) run github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION) \
		-create -ext sql -dir migrations -seq $(NAME)

.PHONY: db-reset
db-reset: ## Полный сброс БД: down -v + up + migrate-up
	docker compose down -v
	docker compose up -d postgres
	@echo ">>> ждём готовности Postgres..."
	@until docker exec liga67-postgres pg_isready -U liga -d liga67 >/dev/null 2>&1; do sleep 0.5; done
	$(MAKE) migrate-up

.PHONY: test-e2e
test-e2e:
	$(GO) test ./internal/delivery/http/e2e/... -v -count=1