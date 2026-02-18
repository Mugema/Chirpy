-- name: CreateUser :one
insert into users(id,created_at,updated_at,email,password) values($1,$2,$3,$4,$5) Returning *;

-- name: GetUserByEmail :one
select * from users where email = $1;

-- name: Reset :exec
delete  from users;

-- name: UpdateUser :one
update users set email = $1, password = $2 ,updated_at = $3 where id = $4 returning *;