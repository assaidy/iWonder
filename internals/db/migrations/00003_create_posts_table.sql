-- +goose Up
-- +goose StatementBegin
create table posts (
    id uuid default gen_random_uuid(),
    user_id uuid not null,
    title varchar(200) not null,
    content varchar not null,
    answered bool not null default false,
    created_at timestamptz not null default now(),

    primary key (id),
    foreign key (user_id) references users (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table posts;
-- +goose StatementEnd
