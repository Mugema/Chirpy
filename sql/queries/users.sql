-- name: CreateUser :one
insert into users(id,created_at,updated_at,email,password,is_chirpy_red) values($1,$2,$3,$4,$5,$6) Returning *;

-- name: GetUserByEmail :one
select * from users where email = $1;

-- name: Reset :exec
delete  from users;

-- name: UpdateUser :one
update users set email = $1, password = $2 ,updated_at = $3 where id = $4 returning *;

-- name: UpgradeUser :exec
update users set is_chirpy_red = $1, updated_at =$2 where id = $3;

