-- +goose Up
-- +goose StatementBegin
create table comments (
    id uuid default gen_random_uuid(),
    post_id uuid not null,
    user_id uuid not null,
    content varchar not null,
    created_at timestamptz not null default now(),

    primary key (id),
    foreign key (user_id) references users (id),
    foreign key (post_id) references posts (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table comments;
-- +goose StatementEnd
