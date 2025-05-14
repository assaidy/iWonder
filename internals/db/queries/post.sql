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
