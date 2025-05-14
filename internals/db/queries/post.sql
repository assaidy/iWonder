-- name: InsertPost :one
insert into posts (id, user_id, title, content)
values ($1, $2, $3, $4)
returning *;

-- name: GetPostByID :one
select * from posts where id = $1;

-- name: CheckPostForUser :one
select exists (select 1 from posts where id = $1 and user_id = $2 for update);

-- name: UpdatePostByID :one
update posts
set
    title = $1,
    content = $2
where id = $3
returning *;

-- name: TogglePostAnswered :one
update posts
set answered = not answered
where id = $1
returning *;

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
