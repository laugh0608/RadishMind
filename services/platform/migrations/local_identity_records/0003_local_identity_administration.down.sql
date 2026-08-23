DROP INDEX IF EXISTS local_workspace_memberships_directory_idx;

ALTER TABLE local_role_assignments
    DROP COLUMN IF EXISTS role_definition_digest,
    DROP COLUMN IF EXISTS role_catalog_version;
