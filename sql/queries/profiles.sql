-- name: CreateProfile :one
INSERT INTO profiles (
    user_id, 
    display_name,
    avatar_url
    )
VALUES (
    sqlc.arg(user_id), 
    sqlc.arg(display_name),
    sqlc.arg(avatar_url)
    )
RETURNING *;

-- name: GetProfileByUserID :one
SELECT * FROM profiles
WHERE user_id = sqlc.arg(user_id);

-- name: UpdateProfile :one
UPDATE profiles
SET display_name = COALESCE(sqlc.narg(display_name), display_name),
    bio           = COALESCE(sqlc.narg(bio), bio),
    social_links  = COALESCE(sqlc.narg(social_links), social_links),
    cosplay_tags  = COALESCE(sqlc.narg(cosplay_tags), cosplay_tags),
    updated_at    = now()
WHERE user_id = sqlc.arg(user_id)
RETURNING *;

-- name: UpdateProfileAvatar :exec
UPDATE profiles
SET avatar_url = sqlc.arg(avatar_url),
    updated_at = now()
WHERE user_id = sqlc.arg(user_id);

-- name: UpdateProfileBanner :exec
UPDATE profiles
SET banner_url = sqlc.arg(banner_url),
    updated_at = now()
WHERE user_id = sqlc.arg(user_id);

-- name: GetUserWithProfileByID :one
-- Gak filter status di sini karena dipakai buat user itu sendiri, bukan publik.
SELECT
    u.id,
    u.username,
    u.status,
    u.is_verified,
    p.display_name,
    p.bio,
    p.avatar_url,
    p.banner_url,
    p.social_links,
    p.cosplay_tags
FROM users u
JOIN profiles p ON p.user_id = u.id
WHERE u.id = sqlc.arg(id);

-- name: GetPublicProfileByUsername :one
-- Buat halaman profil publik (/u/{username}).
-- Filter status = 'ACTIVE' biar profil user yang lagi di-suspend/ban gak muncul.
SELECT
    u.id,
    u.username,
    u.is_verified,
    p.display_name,
    p.bio,
    p.avatar_url,
    p.banner_url,
    p.social_links,
    p.cosplay_tags
FROM users u
JOIN profiles p ON p.user_id = u.id
WHERE u.username = sqlc.arg(username)
  AND u.status = 'ACTIVE'
  OR u.staus = 'ON_BOARDING';