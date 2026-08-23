ALTER TABLE local_role_assignments
    ADD COLUMN role_catalog_version TEXT;

ALTER TABLE local_role_assignments
    ADD COLUMN role_definition_digest TEXT CHECK (
        (role_catalog_version IS NULL AND role_definition_digest IS NULL)
        OR
        (
            role_catalog_version <> ''
            AND length(role_definition_digest) = 71
            AND substr(role_definition_digest, 1, 7) = 'sha256:'
            AND substr(role_definition_digest, 8) NOT GLOB '*[^0-9a-f]*'
        )
    );

CREATE INDEX local_workspace_memberships_directory_idx
    ON local_workspace_memberships(
        tenant_ref,
        workspace_id,
        lifecycle_state,
        updated_at_unix_nano DESC,
        membership_id DESC
    );
