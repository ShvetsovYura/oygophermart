-- +goose Up
-- +goose StatementBegin
alter table "order" drop column order_id;
alter table "order" drop column "type";
alter table "order" drop column "value";
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table "order"  add column "value" bigint NOT NULL;
alter table "order"  add column "order_id" text NOT NULL;
alter table "order" add column "type" text NOT NULL;
-- +goose StatementEnd
