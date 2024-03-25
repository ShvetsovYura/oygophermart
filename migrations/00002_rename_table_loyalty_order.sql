-- +goose Up
-- +goose StatementBegin
ALTER table loyalty_order RENAME TO "order";
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER table "order" RENAME TO loyalty_order;
-- +goose StatementEnd
