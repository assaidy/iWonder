-- +goose Up
-- +goose StatementBegin
create table tags (
    id int generated always as identity,
    name varchar(50) not null unique,
    created_at timestamptz not null default now(),

    primary key (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table tags;
-- +goose StatementEnd
