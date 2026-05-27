-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS subscriptions
(
    id           BIGSERIAL PRIMARY KEY,
    service_name VARCHAR(255) NOT NULL,
    price        INTEGER      NOT NULL CHECK (price >= 0),
    user_id      UUID         NOT NULL,
    start_date   DATE         NOT NULL,
    end_date     DATE,

    CONSTRAINT chk_dates CHECK (end_date IS NULL OR end_date >= start_date)
);

-- Индекс для фильтрации по пользователю + периоду
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_period
    ON subscriptions (user_id, start_date, end_date)
    INCLUDE (price);

-- Индекс для фильтрации по названию подписки + периоду
CREATE INDEX IF NOT EXISTS idx_subscriptions_service_period
    ON subscriptions (service_name, start_date, end_date)
    INCLUDE (price);

-- Составной индекс для фильтрации сразу по пользователю и названию
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_service_period
    ON subscriptions (user_id, service_name, start_date, end_date)
    INCLUDE (price);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_subscriptions_user_service_period;
DROP INDEX IF EXISTS idx_subscriptions_service_period;
DROP INDEX IF EXISTS idx_subscriptions_user_period;
DROP TABLE IF EXISTS subscriptions;

-- +goose StatementEnd