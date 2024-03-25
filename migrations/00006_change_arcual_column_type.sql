-- +goose Up
-- +goose StatementBegin
ALTER TABLE loyalty ALTER COLUMN accrual TYPE double precision;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE loyalty ALTER COLUMN accrual TYPE bigint;
-- +goose StatementEnd
