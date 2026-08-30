-- name: CreatePostLike :one
INSERT INTO post_likes (
    id, user_id, post_id
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetPostLikesWithDetails :many
SELECT 
    pl.id AS like_id,
    pl.created_at AS liked_at,
    u.id AS user_id,
    u.username,
    p.display_name,
    p.avatar_url
FROM post_likes pl
JOIN users u ON pl.user_id = u.id
JOIN profiles p ON u.id = p.user_id
WHERE pl.post_id = $1
ORDER BY pl.created_at DESC;

-- name: DeletePostLike :execrows
DELETE FROM post_likes
WHERE post_id = $1 AND user_id = $2;

-- name: GetUserLikes :many
SELECT 
    id,
    post_id,
    created_at
FROM post_likes
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CheckUserLikedPost :one
SELECT EXISTS (
    SELECT 1 
    FROM post_likes
    WHERE user_id = $1 AND post_id = $2
);

-- name: GetUserLikedPostsInArray :many
SELECT post_id
FROM post_likes
WHERE user_id = $1 AND post_id = ANY($2::uuid[]);