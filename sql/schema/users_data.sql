create type user_status as enum ('ACTIVE', 'SUSPENDED', 'DEACTIVATE', 'ON_BOARDING');
create type user_role as enum ('ADMIN', 'USER');

CREATE TABLE users (
    id UUID PRIMARY KEY REFERENCES users_auth(id) ON DELETE CASCADE, 
    username VARCHAR(30) UNIQUE NOT NULL,
    status user_status not null default 'ON_BOARDING',
    status_reason TEXT,
    status_until TIMESTAMPTZ,        
    role user_role NOT NULL DEFAULT 'USER',
    is_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);