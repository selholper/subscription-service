-- +goose Up
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_subscriptions_user_period;
DROP INDEX IF EXISTS idx_subscriptions_service_period;
ALTER INDEX idx_subscriptions_user_service_period RENAME TO idx_subscriptions_cost_filter;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER INDEX idx_subscriptions_cost_filter RENAME TO idx_subscriptions_user_service_period;
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_period
    ON subscriptions (user_id, start_date, end_date)
    INCLUDE (price);
CREATE INDEX IF NOT EXISTS idx_subscriptions_service_period
    ON subscriptions (service_name, start_date, end_date)
    INCLUDE (price);

-- +goose StatementEnd
