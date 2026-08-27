CREATE TYPE post_visibility AS ENUM('PUBLIC', 'PRIVATE', 'EXCLUSIVE');

CREATE TABLE posts(
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    caption TEXT,
    location varchar(255),
    visibility post_visibility not null default 'PUBLIC',
    like_count      INT NOT NULL DEFAULT 0,
    comment_count   INT NOT NULL DEFAULT 0,
    bookmark_count  INT NOT NULL DEFAULT 0,
    share_count     INT NOT NULL DEFAULT 0,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
