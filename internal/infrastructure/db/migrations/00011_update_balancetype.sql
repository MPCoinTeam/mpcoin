-- +goose Up
-- +goose StatementBegin
ALTER TABLE balances
ALTER COLUMN balance type BIGINT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE balances
ALTER COLUMN balance type NUMERIC;
-- +goose StatementEnd
