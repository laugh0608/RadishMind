CREATE ROLE radishmind_runtime
    LOGIN
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOINHERIT;

REVOKE CREATE ON SCHEMA public FROM PUBLIC;
GRANT CONNECT ON DATABASE radishmind_saved_draft_test TO radishmind_runtime;
GRANT USAGE ON SCHEMA public TO radishmind_runtime;

ALTER DEFAULT PRIVILEGES FOR ROLE radishmind_migrator IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO radishmind_runtime;
