-- name: CreateChirp :one
insert into chirp(
    id,
    user_id,
    created_at,
    updated_at,
    body
) values ($1, $2,$3,$4,$5) returning *;

-- name: GetChirps :many
select * from chirp order by created_at asc;

-- name: GetChirpById :one
select * from chirp where id = $1;

-- name: DeleteChirp :exec
Delete from chirp where id = $1;