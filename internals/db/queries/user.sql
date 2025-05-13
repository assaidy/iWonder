-- name: InsertUser :execrows
insert into users (id, name, bio, username, hashed_password)
values ($1, $2, $3, $4, $5)
on conflict (username) do nothing;

-- name: GetUserByUsername :one
select * from users where username = $1;

-- name: CheckUserID :one
select exists(select 1 from users where id = $1 for update);

-- name: CheckUsername :one
select exists(select 1 from users where username = $1 for update);

-- name: UpdateUserByID :exec
update users
set
    name = $1,
    bio = $2,
    username = $3,
    hashed_password = $4
where id = $5;

-- name: DeleteUserById :exec
delete from users where id = $1;
