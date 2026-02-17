-- name: CreateUser :one
insert into users(id,created_at,updated_at,email,password) values($1,$2,$3,$4,$5) Returning *;

-- name: GetUserByEmail :one
select * from users where email = $1;

-- name: Reset :exec
delete  from users;