-- +goose Up
-- +goose StatementBegin
alter table loyalty add column accrual bigint NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table loyalty drop column accrual;
-- +goose StatementEnd
