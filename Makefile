.PHONY: help migrate-up migrate-down migrate-status migrate-reset migrate-version \
	build run test test-integration swagger tidy fmt vet \
	docker-build compose-up compose-down compose-logs

ENV_FILE 	?= .env
CMD_MIGRATE = go run ./cmd/migrate
CMD_API 	= go run ./cmd/api

help: ## Показать список доступных команд
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

migrate-up: ## Применить все pending миграции
	$(CMD_MIGRATE) -env $(ENV_FILE) -cmd up

migrate-down: ## Откатить последнюю применённую миграцию
	$(CMD_MIGRATE) -env $(ENV_FILE) -cmd down

migrate-status: ## Показать статус всех миграций
	$(CMD_MIGRATE) -env $(ENV_FILE) -cmd status

migrate-reset: ## Откатить все миграции
	$(CMD_MIGRATE) -env $(ENV_FILE) -cmd reset

migrate-version: ## Показать текущую версию миграции
	$(CMD_MIGRATE) -env $(ENV_FILE) -cmd version

build: ## Собрать бинарник API в ./bin/api
	go build -o ./bin/api ./cmd/api

run: ## Запустить API локально
	$(CMD_API) -env $(ENV_FILE)

test: ## Запустить unit-тесты
	go test ./...

test-integration: ## Запустить интеграционные тесты
	go test -tags=integration ./...

swagger: ## Сгенерировать swagger-документацию
	swag init -g cmd/api/main.go -o docs --parseInternal --parseDepth 2

tidy: ## Привести в порядок зависимости
	go mod tidy

fmt: ## Отформатировать код
	gofmt -w cmd internal

vet: ## Запустить статический анализ
	go vet ./...

docker-build: ## Собрать docker-образ сервиса
	docker build -t subscription-service .

compose-up: ## Поднять сервис со всеми зависимостями (db, миграции, api)
	docker compose up --build -d

compose-down: ## Остановить сервис (данные в volume сохраняются)
	docker compose down

compose-logs: ## Показать логи сервиса
	docker compose logs -f api
