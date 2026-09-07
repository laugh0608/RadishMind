ALTER TABLE local_role_assignments
    ADD COLUMN role_catalog_version text,
    ADD COLUMN role_definition_digest text,
    ADD CONSTRAINT local_role_assignments_catalog_metadata_check CHECK (
        (role_catalog_version IS NULL AND role_definition_digest IS NULL)
        OR
        (
            role_catalog_version <> ''
            AND role_definition_digest ~ '^sha256:[0-9a-f]{64}$'
        )
    );

CREATE INDEX local_workspace_memberships_directory_idx
    ON local_workspace_memberships(
        tenant_ref,
        workspace_id,
        lifecycle_state,
        updated_at DESC,
        membership_id DESC
    );
