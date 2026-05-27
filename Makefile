.PHONY: help migrate-up migrate-down migrate-status migrate-reset migrate-version

ENV_FILE 	?= .env
CMD_MIGRATE = go run ./cmd/migrate

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
