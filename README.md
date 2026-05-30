# Subscription Service

REST-сервис для агрегации данных об онлайн-подписках пользователей.

## Возможности

- CRUDL-операции над записями о подписках
- Подсчёт суммарной стоимости подписок за период с фильтрацией по пользователю и названию сервиса
- Swagger-документация
- Структурированное логирование (zap)
- Запуск через docker-compose с сохранением данных PostgreSQL в volume

## Архитектура

Проект следует принципам чистой архитектуры и Go Standard Project Layout:

```
cmd/
  api/            точка входа HTTP-сервиса
  migrate/        запуск миграций (goose)
internal/
  app/            DI-слой: сборка и запуск приложения
  config/         конфигурация из окружения
  database/       подключение к PostgreSQL (GORM)
  domain/         доменная сущность (модель данных / DAO)
  dto/            модели запросов/ответов HTTP
  logger/         инициализация zap-логгера
  repository/     интерфейс репозитория
    postgres/     реализация репозитория на GORM
  service/        бизнес-логика (интерфейс + реализация)
  transport/http/ HTTP-хендлеры, роутинг, middleware
docs/             сгенерированная swagger-документация
migrations/       SQL-миграции goose
```

Зависимости направлены внутрь: `handler → service → repository`, связи между слоями
осуществляются через интерфейсы, сборка зависимостей — в пакете `internal/app`.

## Требования

- Go 1.25+
- Docker и Docker Compose (для запуска через контейнеры)
- `swag` CLI для регенерации документации: `go install github.com/swaggo/swag/cmd/swag@latest`

## Конфигурация

Скопируйте `.env.example` в `.env` и при необходимости измените значения:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=mypassword
DB_NAME=mydb
DB_SSLMODE=disable

HTTP_PORT=8080
LOG_LEVEL=info
```

## Запуск через docker-compose (рекомендуется)

Поднимает PostgreSQL, применяет миграции и запускает API:

```bash
make compose-up
```

Сервис будет доступен на `http://localhost:8080`, Swagger UI — на
`http://localhost:8080/swagger/index.html`.

Остановить (данные PostgreSQL сохраняются в volume `pgdata`):

```bash
make compose-down
```

Логи API:

```bash
make compose-logs
```

## Локальный запуск

1. Поднимите PostgreSQL (например, через docker-compose только для базы) и пропишите
   доступы в `.env`.
2. Примените миграции:

   ```bash
   make migrate-up
   ```

3. Запустите сервис:

   ```bash
   make run
   ```

## API

Базовый URL: `http://localhost:8080`

| Метод  | Путь                    | Описание                                  |
|--------|-------------------------|-------------------------------------------|
| POST   | `/subscriptions`        | Создать подписку                          |
| GET    | `/subscriptions/{id}`   | Получить подписку по ID                   |
| PUT    | `/subscriptions/{id}`   | Обновить подписку                         |
| DELETE | `/subscriptions/{id}`   | Удалить подписку                          |
| GET    | `/subscriptions`        | Список подписок (фильтры + пагинация)     |
| GET    | `/subscriptions/cost`   | Суммарная стоимость за период             |
| GET    | `/healthz`              | Проверка работоспособности                |

Даты (`start_date`, `end_date`, а также `from`/`to` в подсчёте стоимости) передаются в
формате `MM-YYYY` (например `07-2025`) и сохраняются как первый день месяца. Стоимость
подписки — целое число рублей.

### Примеры

Создание подписки:

```bash
curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

Суммарная стоимость за период с фильтрацией:

```bash
curl "http://localhost:8080/subscriptions/cost?from=01-2025&to=12-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus"
```

## Swagger

Документация доступна на `http://localhost:8080/swagger/index.html` после запуска сервиса.

Регенерация после изменения аннотаций:

```bash
make swagger
```

## Тесты

Unit-тесты:

```bash
make test
```

Интеграционные тесты (полный HTTP-стек через `httptest`):

```bash
make test-integration
```

## Makefile

Полный список команд:

```bash
make help
```
