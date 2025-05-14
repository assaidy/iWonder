-- +goose Up
-- +goose StatementBegin
create table post_tags (
    post_id uuid,
    tag_id int,

    primary key (post_id, tag_id),
    foreign key (post_id) references posts (id) on delete cascade,
    foreign key (tag_id) references tags (id) on delete cascade
);

create index on post_tags(post_id);
create index on post_tags(tag_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table post_tags;
-- +goose StatementEnd
