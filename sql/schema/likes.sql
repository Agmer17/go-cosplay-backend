CREATE TABLE post_likes(
    id UUID primary key,
    user_id UUID NOT NULL REFERENCES users(id) on DELETE CASCADE,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    created_at timestamptz not null default now(),
    UNIQUE (post_id, user_id)
);