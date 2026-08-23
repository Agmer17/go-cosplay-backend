create type auth_provider as enum ('GOOGLE', 'DISCORD');


CREATE TABLE users_auth (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider auth_provider NOT NULL,
    provider_openid varchar(255) NOT NULL,
    email varchar(255) not null,
    created_at timestamp with time zone DEFAULT now(),

    CONSTRAINT users_auth_provider_openid_key UNIQUE (provider, provider_openid),
    constraint users_email_provider unique (email, provider)
);
