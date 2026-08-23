-- name: CreateUserAuth :one
INSERT INTO users_auth (
    provider,
    provider_openid,
    email
)
VALUES (
    sqlc.arg(provider),
    sqlc.arg(provider_openid),
    sqlc.arg(email)
)
RETURNING *;

-- name: GetUserAuthByID :one
SELECT
    id,
    provider,
    provider_openid,
    email,
    created_at
FROM users_auth
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetUserAuthByProviderOpenID :one
SELECT
    *
FROM users_auth
WHERE provider = sqlc.arg(provider)
  AND provider_openid = sqlc.arg(provider_openid)
LIMIT 1;

-- name: DeleteUserAuth :exec
DELETE FROM users_auth
WHERE id = sqlc.arg(id);