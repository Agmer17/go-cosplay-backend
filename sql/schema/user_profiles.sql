CREATE TABLE profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    display_name  VARCHAR(75),
    bio           TEXT,
    avatar_url    TEXT,
    banner_url    TEXT,
    social_links  JSONB,      -- {"twitter": "...", "ig": "..."}
    cosplay_tags  TEXT[],     -- fandom/genre yg biasa dicosplay
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);