-- Migration DOWN: 001_create_users
-- Drops the users table, restoring the exact schema
-- that existed before this migration.

DROP TABLE IF EXISTS users;
