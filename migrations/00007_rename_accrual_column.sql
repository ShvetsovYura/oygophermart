-- +goose Up
-- +goose StatementBegin
ALTER TABLE loyalty RENAME COLUMN accrual TO "value";
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE loyalty RENAME COLUMN "value" TO accrual;
-- +goose StatementEnd
