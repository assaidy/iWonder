-- +goose Up
-- +goose StatementBegin
		/*
post_answer:
- post_id
- comment_id
pk(post_id, comment_id)
*/
create table posts_answer (
    post_id uuid,
    comment_id uuid,

    primary key (post_id, comment_id),
    foreign key (post_id) references posts (id),
    foreign key (comment_id) references comments (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table posts_answer;
-- +goose StatementEnd
