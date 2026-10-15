SHELL := /bin/bash
GO              ?= go
MOCKERY_VERSION ?= v3.8.0
GOLANGCI_VERSION ?= v1.59.1

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

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

.PHONY: test-cover
test-cover: ## Покрытие + HTML
	$(GO) test ./... -race -count=1 -coverprofile=coverage.out -covermode=atomic
	$(GO) tool cover -html=coverage.out -o coverage.html

.PHONY: test-domain
test-domain: ## Только domain
	$(GO) test ./internal/domain/... -race -count=1 -v

## --- Качество ----------------------------------------------------------

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: check
check: fmt vet lint test ## Локальный чек перед PR

## --- Сборка / очистка --------------------------------------------------

.PHONY: build
build:
	$(GO) build ./...

.PHONY: clean
clean:
	rm -rf internal/testsupport/mocks coverage.out coverage.html