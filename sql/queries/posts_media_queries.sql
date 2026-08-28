-- name: CreatePostMedia :one
INSERT INTO posts_media (
    post_id, media_type, media_url, display_order
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: CreatePostMediaBatch :many
INSERT INTO posts_media (post_id, media_type, media_url, display_order)
SELECT
  $1,
  unnest(sqlc.arg(media_types)::text[])::post_media_type,
  unnest(sqlc.arg(media_urls)::text[]),
  unnest(sqlc.arg(display_orders)::smallint[])
RETURNING *;

-- name: GetMediaByPostID :many
SELECT * FROM posts_media
WHERE post_id = $1
ORDER BY display_order;

-- name: DeletePostMediaByPostID :exec
-- Berguna buat flow edit post: hapus semua media lama,
-- lalu CreatePostMediaBatch lagi dengan media yang baru
DELETE FROM posts_media WHERE post_id = $1;