-- name: InsertPost :one
insert into posts (id, user_id, title, content)
values ($1, $2, $3, $4)
returning *;

-- name: GetPostByID :one
select * from posts where id = $1;

-- name: CheckPostForUser :one
select exists (select 1 from posts where id = $1 and user_id = $2 for update);

-- name: UpdatePostByID :exec
update posts
set
    title = $1,
    content = $2
where id = $3;

-- name: DeletePostByID :exec
delete from posts where id = $1;

-- name: InsertTag :one
insert into tags (name) 
values ($1)
on conflict (name) do update 
    set name = excluded.name 
returning id;

-- name: InsertTagForPost :exec
insert into post_tags (post_id, tag_id)
values ($1, $2);

-- name: DeleteTagForPost :exec
delete from post_tags
where post_id = $1 and tag_id = (select id from tags where name = $2);

-- name: GetPostTags :many
select t.name
from tags t
join post_tags pt on pt.tag_id = t.id
where pt.post_id = $1;

-- name: CheckPost :one
select exists (select 1 from posts where id = $1 for update);

-- name: InsertComment :exec
insert into comments (id, post_id, user_id, content)
values ($1, $2, $3, $4);

-- name: CheckCommentForUser :one
select exists (select 1 from comments where id = $1 and user_id = $2 for update);

-- name: UpdateComment :exec
update comments
set content = $1
where id = $2;

-- name: DeleteComment :exec
delete from comments where id = $1;

-- name: GetPostComments :many
select * from comments
where 
    post_id = $1 and
    created_at <= coalesce(
        nullif(sqlc.arg(created_at)::timestamptz, '0001-01-01 00:00:00'::timestamptz),
        now()::timestamptz
    )
order by created_at desc
limit $2;

-- name: CheckComment :one
select exists (select 1 from comments where id = $1 for update);

-- name: InsertCommentVote :exec
insert into comment_votes (comment_id, user_id, kind)
values ($1, $2, $3)
on conflict (comment_id, user_id) do update
    set kind = excluded.kind;

-- name: CheckCommentVoteForUser :one
select exists (select 1 from comment_votes where comment_id = $1 and user_id = $2 for update);

-- name: DeleteCommentVote :exec
delete from comment_votes where comment_id = $1 and user_id = $2;

-- name: InsertPostAnswer :exec
insert into posts_answer (post_id, comment_id)
values ($1, $2)
on conflict (post_id) do update
    set comment_id = excluded.comment_id;

-- name: DeletePostAnswer :exec
delete from posts_answer where post_id = $1;
