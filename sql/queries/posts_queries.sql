-- name: CreatePost :one
INSERT INTO posts (
    id, user_id, caption, location, visibility
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;


-- name: GetPostByID :one
SELECT
    p.*,
    COALESCE(
        jsonb_agg(pm.* ORDER BY pm.display_order)
            FILTER (WHERE pm.id IS NOT NULL),
        '[]'::jsonb
    )::jsonb AS media,
    false as is_liked
FROM posts p
LEFT JOIN posts_media pm ON pm.post_id = p.id
WHERE p.id = $1
GROUP BY p.id;

-- name: GetPostsByIDArray :many
SELECT
    p.*,
    COALESCE(
        jsonb_agg(pm.* ORDER BY pm.display_order)
            FILTER (WHERE pm.id IS NOT NULL),
        '[]'::jsonb
    )::jsonb AS media,
    false as is_liked
FROM posts p
LEFT JOIN posts_media pm ON pm.post_id = p.id
WHERE p.id = ANY($1::uuid[])
GROUP BY p.id;

-- name: GetPostsDataOnlyById :one
SELECT * FROM posts where id = sqlc.arg(id);

-- name: ListPostsByUser :many
-- Keyset pagination pakai created_at + id sebagai tie-breaker
-- (lebih stabil daripada OFFSET buat feed yang terus nambah data baru)
SELECT
    p.*,
    COALESCE(
        jsonb_agg(pm.* ORDER BY pm.display_order) FILTER (WHERE pm.id IS NOT NULL),
        '[]'::jsonb
    ) AS media
FROM posts p
LEFT JOIN posts_media pm ON pm.post_id = p.id
WHERE p.user_id = $1
    AND (p.created_at, p.id) < (sqlc.arg(cursor_created_at), sqlc.arg(cursor_id))
GROUP BY p.id
ORDER BY p.created_at DESC, p.id DESC
LIMIT $2;

-- name: UpdatePost :one
UPDATE posts
SET
    caption    = COALESCE(sqlc.narg(caption), caption),
    location   =  COALESCE(sqlc.narg(location), location),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1;

-- name: IncrementPostLikeCount :exec
UPDATE posts SET like_count = like_count + 1 WHERE id = $1;

-- name: DecrementPostLikeCount :exec
UPDATE posts SET like_count = GREATEST(like_count - 1, 0) WHERE id = $1;

-- name: IncrementPostCommentCount :exec
UPDATE posts SET comment_count = comment_count + 1 WHERE id = $1;

-- name: DecrementPostCommentCount :exec
UPDATE posts SET comment_count = GREATEST(comment_count - 1, 0) WHERE id = $1;

-- name: IncrementPostBookmarkCount :exec
UPDATE posts SET bookmark_count = bookmark_count + 1 WHERE id = $1;

-- name: DecrementPostBookmarkCount :exec
UPDATE posts SET bookmark_count = GREATEST(bookmark_count - 1, 0) WHERE id = $1;

-- name: IncrementPostShareCount :one
UPDATE posts 
SET share_count = share_count + 1 
WHERE id = $1 
RETURNING share_count;

-- name: DecrementPostShareCount :one
UPDATE posts 
SET share_count = GREATEST(share_count - 1, 0) 
WHERE id = $1 
RETURNING share_count;