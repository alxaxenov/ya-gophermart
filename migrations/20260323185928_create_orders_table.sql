-- +goose Up
SELECT 'up SQL query';
CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE orders (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL,
    number VARCHAR(50) NOT NULL UNIQUE,
    accrual DECIMAL(10,1) DEFAULT NULL,
    status order_status NOT NULL DEFAULT 'NEW',
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_status_uploaded ON orders (status, uploaded_at)

-- +goose Down
SELECT 'down SQL query';
DROP TABLE IF EXISTS orders;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_status_uploaded;
