-- Migration UP: 001_create_users
-- Users for Task Manager auth. Passwords are stored only as
-- bcrypt hashes in password_hash, never in plain text.

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_users_name_not_empty
        CHECK (char_length(btrim(name)) > 0),
    CONSTRAINT chk_users_email_not_empty
        CHECK (char_length(btrim(email)) > 0)
);
