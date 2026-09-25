-- 0001_init.sql
-- Initial schema for the example application.
-- This migration is applied automatically at startup by database.Migrate.

CREATE TABLE example_table (
    id   INTEGER NOT NULL PRIMARY KEY,
    name TEXT
);
