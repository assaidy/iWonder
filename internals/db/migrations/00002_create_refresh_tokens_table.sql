-- +goose Up
-- +goose StatementBegin
create table refresh_tokens (
    token varchar(100),
    user_id uuid not null,
    created_at timestamptz not null default now(),
    expires_at timestamptz not null,

    primary key (token),
    foreign key (user_id) references users (id) on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table refresh_tokens;
-- +goose StatementEnd
