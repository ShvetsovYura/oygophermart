-- +goose Up
-- +goose StatementBegin
create table if not exists loyalty
(
	id bigserial not null,
	order_id text not null,
	created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
	constraint loyalty_pkey primary key(id),
	constraint order_fk foreign key (order_id) references "order"("id")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists "loyalty";
-- +goose StatementEnd
