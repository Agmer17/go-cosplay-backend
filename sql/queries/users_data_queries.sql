-- name: CreateUser :one
-- Dipanggil setelah users_auth dibuat & user pilih username (step onboarding).
INSERT INTO users (
    id, 
    username,
    role
    )
VALUES (
    sqlc.arg(id), 
    sqlc.arg(username),
    sqlc.arg(role)
    )
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = sqlc.arg(id);

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = sqlc.arg(username);

-- name: IsUsernameTaken :one
-- Buat live-check saat user ngetik username di form onboarding.
SELECT EXISTS (
    SELECT 1 FROM users WHERE username = sqlc.arg(username)
) AS taken;

-- name: UpdateUsername :one
UPDATE users
SET username = sqlc.arg(username),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateUserStatus :one
-- Dipanggil dari moderation service tiap kali ada moderation_action baru
-- (suspend/ban/reinstate). status_reason & status_until nullable karena
-- 'reinstate' biasanya gak butuh keduanya.
UPDATE users
SET status = sqlc.arg(status),
    status_reason = sqlc.narg(status_reason),
    status_until = sqlc.narg(status_until),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetUserVerified :exec
UPDATE users
SET is_verified = sqlc.arg(is_verified),
    updated_at = now()
WHERE id = sqlc.arg(id);

-- name: ListUsersByStatus :many
-- Buat admin dashboard, misal liat semua user yang lagi 'SUSPENDED'.
SELECT * FROM users
WHERE status = sqlc.arg(status)
ORDER BY created_at DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: ListExpiredSuspensions :many
-- Buat cron/background job: cari user yang status_until-nya udah lewat
-- biar bisa di-auto-reactivate balik ke 'ACTIVE'.
SELECT * FROM users
WHERE status != 'ACTIVE'
  AND status_until IS NOT NULL
  AND status_until <= now();