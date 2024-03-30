-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public."user"
(
    id bigserial NOT NULL,
    login text NOT NULL,
    pwd_hash text NOT NULL,
    CONSTRAINT user_pkey PRIMARY KEY (id),
    CONSTRAINT unique_login UNIQUE (login)
);

CREATE TABLE IF NOT EXISTS public.loyalty_order
(
    id text NOT NULL,
    order_id text NOT NULL,
    type text NOT NULL,
    status text NOT NULL,
    value bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT loyalty_order_pkey PRIMARY KEY (id),
    CONSTRAINT user_fk FOREIGN KEY (user_id)
        REFERENCES public."user" (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.loyalty_order;
DROP TABLE IF EXISTS public."user";
-- +goose StatementEnd
