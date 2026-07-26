CREATE TABLE agent_copilot_profile_drafts (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (application_id ~ '^app_[a-z2-7]{16}$'),
    owner_subject_ref text NOT NULL CHECK (btrim(owner_subject_ref) <> ''),
    profile_id text NOT NULL CHECK (profile_id ~ '^acpf_[a-z2-7]{16}$'),
    draft_version bigint NOT NULL CHECK (draft_version > 0),
    profile_digest text NOT NULL CHECK (profile_digest ~ '^sha256:[a-f0-9]{64}$'),
    policy_digest text NOT NULL CHECK (policy_digest ~ '^sha256:[a-f0-9]{64}$'),
    updated_at timestamptz NOT NULL,
    sanitized_draft_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_draft_payload) = 'object'
        AND sanitized_draft_payload->>'schema_version' = 'agent_copilot_profile_draft.v1'
        AND sanitized_draft_payload->>'tenant_ref' = tenant_ref
        AND sanitized_draft_payload->>'workspace_id' = workspace_id
        AND sanitized_draft_payload->>'application_id' = application_id
        AND sanitized_draft_payload->>'owner_subject_ref' = owner_subject_ref
        AND sanitized_draft_payload->>'profile_id' = profile_id
        AND (sanitized_draft_payload->>'draft_version')::bigint = draft_version
        AND sanitized_draft_payload->>'profile_digest' = profile_digest
        AND sanitized_draft_payload->>'policy_digest' = policy_digest
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, profile_id)
);

CREATE TABLE agent_copilot_profile_versions (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    owner_subject_ref text NOT NULL,
    profile_id text NOT NULL,
    profile_version bigint NOT NULL CHECK (profile_version > 0),
    source_draft_version bigint NOT NULL CHECK (source_draft_version > 0),
    profile_digest text NOT NULL CHECK (profile_digest ~ '^sha256:[a-f0-9]{64}$'),
    policy_digest text NOT NULL CHECK (policy_digest ~ '^sha256:[a-f0-9]{64}$'),
    published_at timestamptz NOT NULL,
    immutable_version_payload jsonb NOT NULL CHECK (
        jsonb_typeof(immutable_version_payload) = 'object'
        AND immutable_version_payload->>'schema_version' = 'agent_copilot_profile_version.v1'
        AND immutable_version_payload->>'tenant_ref' = tenant_ref
        AND immutable_version_payload->>'workspace_id' = workspace_id
        AND immutable_version_payload->>'application_id' = application_id
        AND immutable_version_payload->>'owner_subject_ref' = owner_subject_ref
        AND immutable_version_payload->>'profile_id' = profile_id
        AND (immutable_version_payload->>'profile_version')::bigint = profile_version
        AND (immutable_version_payload->>'source_draft_version')::bigint = source_draft_version
        AND immutable_version_payload->>'profile_digest' = profile_digest
        AND immutable_version_payload->>'policy_digest' = policy_digest
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, profile_id, profile_version),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, profile_id, source_draft_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, profile_id)
        REFERENCES agent_copilot_profile_drafts (tenant_ref, workspace_id, application_id, owner_subject_ref, profile_id)
        ON DELETE RESTRICT
);

CREATE INDEX agent_copilot_profile_drafts_scope_idx ON agent_copilot_profile_drafts
    (tenant_ref, workspace_id, application_id, owner_subject_ref, updated_at DESC, profile_id ASC);
CREATE INDEX agent_copilot_profile_versions_scope_idx ON agent_copilot_profile_versions
    (tenant_ref, workspace_id, application_id, owner_subject_ref, profile_id, profile_version DESC);

CREATE FUNCTION enforce_agent_copilot_profile_draft_update() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id OR NEW.application_id <> OLD.application_id
       OR NEW.owner_subject_ref <> OLD.owner_subject_ref OR NEW.profile_id <> OLD.profile_id
       OR NEW.draft_version <> OLD.draft_version + 1
       OR NEW.sanitized_draft_payload->>'created_at' <> OLD.sanitized_draft_payload->>'created_at'
       OR NEW.sanitized_draft_payload->>'created_by_actor_ref' <> OLD.sanitized_draft_payload->>'created_by_actor_ref' THEN
        RAISE EXCEPTION 'agent copilot profile draft transition is invalid';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER agent_copilot_profile_drafts_controlled_update
BEFORE UPDATE ON agent_copilot_profile_drafts FOR EACH ROW
EXECUTE FUNCTION enforce_agent_copilot_profile_draft_update();

CREATE FUNCTION reject_agent_copilot_profile_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'agent copilot profile resource is immutable';
END;
$$;
CREATE TRIGGER agent_copilot_profile_drafts_no_delete
BEFORE DELETE ON agent_copilot_profile_drafts FOR EACH ROW
EXECUTE FUNCTION reject_agent_copilot_profile_mutation();
CREATE TRIGGER agent_copilot_profile_versions_no_update
BEFORE UPDATE ON agent_copilot_profile_versions FOR EACH ROW
EXECUTE FUNCTION reject_agent_copilot_profile_mutation();
CREATE TRIGGER agent_copilot_profile_versions_no_delete
BEFORE DELETE ON agent_copilot_profile_versions FOR EACH ROW
EXECUTE FUNCTION reject_agent_copilot_profile_mutation();
