create type auth_provider as enum ('GOOGLE', 'DISCORD');
create type user_role as enum ('ADMIN', 'USER');

CREATE TABLE users_auth (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider auth_provider NOT NULL,
    provider_openid varchar(255) NOT NULL,
    email varchar(255) not null,
    role user_role not null default 'USER',
    created_at timestamp with time zone DEFAULT now(),

    -- Constraint 1: Komrebinasi provider dan openid harus unik (beda provider boleh punya openid yang sama)
    CONSTRAINT users_auth_provider_openid_key UNIQUE (provider, provider_openid),
    constraint users_email_provider unique (email, provider)
);
