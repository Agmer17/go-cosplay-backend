CREATE TYPE post_media_type AS ENUM ('IMAGE', 'VIDEO');
CREATE TABLE posts_media(
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id           UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    media_type        post_media_type NOT NULL,
    media_url         TEXT NOT NULL,
    display_order     SMALLINT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (post_id, display_order)
)