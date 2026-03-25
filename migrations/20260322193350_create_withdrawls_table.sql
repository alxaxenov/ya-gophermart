-- +goose Up
SELECT 'up SQL query';
CREATE TABLE withdrawals (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL,
    order_number VARCHAR(50) NOT NULL UNIQUE,
    sum DECIMAL(10,2) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_withdrawals_user_id ON withdrawals (user_id);



-- +goose Down
SELECT 'down SQL query';
DROP TABLE IF EXISTS withdrawals;
DROP INDEX IF EXISTS idx_withdrawals_user_id;
