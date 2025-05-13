-- name: InsertRefreshToken :exec
insert into refresh_tokens (token, user_id, expires_at)
values ($1, $2, $3);
