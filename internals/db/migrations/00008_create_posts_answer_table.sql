-- +goose Up
-- +goose StatementBegin
create table post_answers (
    post_id uuid,
    comment_id uuid not null,

    primary key (post_id),
    foreign key (post_id) references posts (id),
    foreign key (comment_id) references comments (id)
);
-- +goose StatementEnd

-- +goose StatementBegin
create function set_post_as_answered()
returns trigger
as $$
begin
    update posts set answered = true;
    return null;
end;
$$ language plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
create trigger trg_set_post_as_answered
after insert or update 
on post_answers for each row
execute function set_post_as_answered();
-- +goose StatementEnd

-- +goose StatementBegin
create function set_post_as_not_answered()
returns trigger
as $$
begin
    update posts set answered = false;
    return null;
end;
$$ language plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
create trigger trg_set_post_as_not_answered
after delete
on post_answers for each row
execute function set_post_as_not_answered();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table post_answers;
drop function set_post_as_answered;
drop function set_post_as_not_answered;
-- +goose StatementEnd
