-- name: CreateToken :one
insert into refresh_tokens(
    token,
    created_at,
    updated_at,
    revoked_at,
    expires_at,
    user_id) values ($1,$2,$3,$4,$5,$6)
    returning *;

-- name: GetToken :one
Select * from refresh_tokens where token = $1;

-- name: RevokeToken :exec
Update refresh_tokens set revoked_at = $1,updated_at = $2 where token = $3;