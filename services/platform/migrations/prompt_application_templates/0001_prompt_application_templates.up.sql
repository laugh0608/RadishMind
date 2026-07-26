CREATE TABLE prompt_application_template_drafts (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (application_id ~ '^app_[a-z2-7]{16}$'),
    owner_subject_ref text NOT NULL CHECK (btrim(owner_subject_ref) <> ''),
    template_id text NOT NULL CHECK (template_id ~ '^ptpl_[a-z2-7]{16}$'),
    draft_version bigint NOT NULL CHECK (draft_version > 0),
    template_digest text NOT NULL CHECK (template_digest ~ '^sha256:[a-f0-9]{64}$'),
    updated_at timestamptz NOT NULL,
    sanitized_draft_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_draft_payload) = 'object'
        AND sanitized_draft_payload->>'schema_version' = 'prompt_application_template_draft.v1'
        AND sanitized_draft_payload->>'tenant_ref' = tenant_ref
        AND sanitized_draft_payload->>'workspace_id' = workspace_id
        AND sanitized_draft_payload->>'application_id' = application_id
        AND sanitized_draft_payload->>'owner_subject_ref' = owner_subject_ref
        AND sanitized_draft_payload->>'template_id' = template_id
        AND (sanitized_draft_payload->>'draft_version')::bigint = draft_version
        AND sanitized_draft_payload->>'template_digest' = template_digest
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, template_id)
);

CREATE TABLE prompt_application_template_versions (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    owner_subject_ref text NOT NULL,
    template_id text NOT NULL,
    template_version bigint NOT NULL CHECK (template_version > 0),
    source_draft_version bigint NOT NULL CHECK (source_draft_version > 0),
    template_digest text NOT NULL CHECK (template_digest ~ '^sha256:[a-f0-9]{64}$'),
    created_at timestamptz NOT NULL,
    immutable_version_payload jsonb NOT NULL CHECK (
        jsonb_typeof(immutable_version_payload) = 'object'
        AND immutable_version_payload->>'schema_version' = 'prompt_application_template_version.v1'
        AND immutable_version_payload->>'tenant_ref' = tenant_ref
        AND immutable_version_payload->>'workspace_id' = workspace_id
        AND immutable_version_payload->>'application_id' = application_id
        AND immutable_version_payload->>'owner_subject_ref' = owner_subject_ref
        AND immutable_version_payload->>'template_id' = template_id
        AND (immutable_version_payload->>'template_version')::bigint = template_version
        AND (immutable_version_payload->>'source_draft_version')::bigint = source_draft_version
        AND immutable_version_payload->>'template_digest' = template_digest
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, template_id, template_version),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, template_id, source_draft_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, template_id)
        REFERENCES prompt_application_template_drafts (tenant_ref, workspace_id, application_id, owner_subject_ref, template_id)
        ON DELETE RESTRICT
);

CREATE INDEX prompt_application_template_drafts_scope_idx ON prompt_application_template_drafts
    (tenant_ref, workspace_id, application_id, owner_subject_ref, updated_at DESC, template_id ASC);
CREATE INDEX prompt_application_template_versions_scope_idx ON prompt_application_template_versions
    (tenant_ref, workspace_id, application_id, owner_subject_ref, template_id, template_version DESC);

CREATE FUNCTION enforce_prompt_application_template_draft_update() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id OR NEW.application_id <> OLD.application_id
       OR NEW.owner_subject_ref <> OLD.owner_subject_ref OR NEW.template_id <> OLD.template_id
       OR NEW.draft_version <> OLD.draft_version + 1
       OR NEW.sanitized_draft_payload->>'created_at' <> OLD.sanitized_draft_payload->>'created_at'
       OR NEW.sanitized_draft_payload->>'created_by_actor_ref' <> OLD.sanitized_draft_payload->>'created_by_actor_ref' THEN
        RAISE EXCEPTION 'prompt application template draft transition is invalid';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER prompt_application_template_drafts_controlled_update
BEFORE UPDATE ON prompt_application_template_drafts FOR EACH ROW
EXECUTE FUNCTION enforce_prompt_application_template_draft_update();

CREATE FUNCTION reject_prompt_application_template_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'prompt application template resource is immutable';
END;
$$;
CREATE TRIGGER prompt_application_template_drafts_no_delete
BEFORE DELETE ON prompt_application_template_drafts FOR EACH ROW
EXECUTE FUNCTION reject_prompt_application_template_mutation();
CREATE TRIGGER prompt_application_template_versions_no_update
BEFORE UPDATE ON prompt_application_template_versions FOR EACH ROW
EXECUTE FUNCTION reject_prompt_application_template_mutation();
CREATE TRIGGER prompt_application_template_versions_no_delete
BEFORE DELETE ON prompt_application_template_versions FOR EACH ROW
EXECUTE FUNCTION reject_prompt_application_template_mutation();
