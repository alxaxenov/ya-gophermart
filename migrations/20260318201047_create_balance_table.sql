-- +goose Up
SELECT 'up SQL query';
CREATE TABLE balance (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL UNIQUE,
    current DECIMAL(10,2) NOT NULL DEFAULT 0,
    withdrawn DECIMAL(10,2) NOT NULL DEFAULT 0
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE IF EXISTS balance;
