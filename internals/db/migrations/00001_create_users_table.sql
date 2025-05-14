-- +goose Up
-- +goose StatementBegin
create table users (
    id uuid default gen_random_uuid(),
    name varchar(100) not null,
    bio varchar,
    username varchar(50) not null unique,
    hashed_password varchar not null,
    created_at timestamptz not null default now(),

    primary key (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table users;
-- +goose StatementEnd
