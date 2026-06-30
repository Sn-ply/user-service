CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT        NOT NULL UNIQUE,
    username        TEXT        NOT NULL UNIQUE,
    password_hash   TEXT        NOT NULL,
    bio             TEXT        NOT NULL DEFAULT '',
    avatar_url      TEXT        NOT NULL DEFAULT '',
    post_count      INT         NOT NULL DEFAULT 0,
    follower_count  INT         NOT NULL DEFAULT 0,
    following_count INT         NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deactivated_at  TIMESTAMPTZ
);

CREATE INDEX idx_users_username ON users (username);
CREATE INDEX idx_users_email    ON users (email);
CREATE INDEX idx_users_created_at_id ON users (created_at ASC, id ASC);
