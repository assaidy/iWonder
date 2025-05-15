-- +goose Up
-- +goose StatementBegin
create table comment_votes (
    comment_id uuid,
    user_id uuid,
    kind varchar(10) not null check (kind in ('up', 'down')),

    primary key (comment_id, user_id),
    foreign key (comment_id) references comments (id),
    foreign key (user_id) references users (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table comment_votes;
-- +goose StatementEnd
