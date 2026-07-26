CREATE TABLE prompt_application_template_drafts (
    tenant_ref TEXT NOT NULL CHECK (length(trim(tenant_ref)) > 0),
    workspace_id TEXT NOT NULL CHECK (length(trim(workspace_id)) > 0),
    application_id TEXT NOT NULL CHECK (application_id GLOB 'app_[a-z2-7]*' AND length(application_id) = 20 AND substr(application_id, 5) NOT GLOB '*[^a-z2-7]*'),
    owner_subject_ref TEXT NOT NULL CHECK (length(trim(owner_subject_ref)) > 0),
    template_id TEXT NOT NULL CHECK (template_id GLOB 'ptpl_[a-z2-7]*' AND length(template_id) = 21 AND substr(template_id, 6) NOT GLOB '*[^a-z2-7]*'),
    draft_version INTEGER NOT NULL CHECK (draft_version > 0),
    template_digest TEXT NOT NULL CHECK (template_digest GLOB 'sha256:[a-f0-9]*' AND length(template_digest) = 71 AND substr(template_digest, 8) NOT GLOB '*[^a-f0-9]*'),
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano > 0),
    sanitized_draft_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_draft_payload)
        AND json_type(sanitized_draft_payload) = 'object'
        AND json_extract(sanitized_draft_payload, '$.schema_version') = 'prompt_application_template_draft.v1'
        AND json_extract(sanitized_draft_payload, '$.tenant_ref') = tenant_ref
        AND json_extract(sanitized_draft_payload, '$.workspace_id') = workspace_id
        AND json_extract(sanitized_draft_payload, '$.application_id') = application_id
        AND json_extract(sanitized_draft_payload, '$.owner_subject_ref') = owner_subject_ref
        AND json_extract(sanitized_draft_payload, '$.template_id') = template_id
        AND json_extract(sanitized_draft_payload, '$.draft_version') = draft_version
        AND json_extract(sanitized_draft_payload, '$.template_digest') = template_digest
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, template_id)
) STRICT;

CREATE TABLE prompt_application_template_versions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    template_id TEXT NOT NULL,
    template_version INTEGER NOT NULL CHECK (template_version > 0),
    source_draft_version INTEGER NOT NULL CHECK (source_draft_version > 0),
    template_digest TEXT NOT NULL CHECK (template_digest GLOB 'sha256:[a-f0-9]*' AND length(template_digest) = 71 AND substr(template_digest, 8) NOT GLOB '*[^a-f0-9]*'),
    created_at_unix_nano INTEGER NOT NULL CHECK (created_at_unix_nano > 0),
    immutable_version_payload TEXT NOT NULL CHECK (
        json_valid(immutable_version_payload)
        AND json_type(immutable_version_payload) = 'object'
        AND json_extract(immutable_version_payload, '$.schema_version') = 'prompt_application_template_version.v1'
        AND json_extract(immutable_version_payload, '$.tenant_ref') = tenant_ref
        AND json_extract(immutable_version_payload, '$.workspace_id') = workspace_id
        AND json_extract(immutable_version_payload, '$.application_id') = application_id
        AND json_extract(immutable_version_payload, '$.owner_subject_ref') = owner_subject_ref
        AND json_extract(immutable_version_payload, '$.template_id') = template_id
        AND json_extract(immutable_version_payload, '$.template_version') = template_version
        AND json_extract(immutable_version_payload, '$.source_draft_version') = source_draft_version
        AND json_extract(immutable_version_payload, '$.template_digest') = template_digest
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, template_id, template_version),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, template_id, source_draft_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, template_id)
        REFERENCES prompt_application_template_drafts (tenant_ref, workspace_id, application_id, owner_subject_ref, template_id)
        ON DELETE RESTRICT
) STRICT;

CREATE INDEX prompt_application_template_drafts_scope_idx ON prompt_application_template_drafts
    (tenant_ref, workspace_id, application_id, owner_subject_ref, updated_at_unix_nano DESC, template_id ASC);
CREATE INDEX prompt_application_template_versions_scope_idx ON prompt_application_template_versions
    (tenant_ref, workspace_id, application_id, owner_subject_ref, template_id, template_version DESC);

CREATE TRIGGER prompt_application_template_drafts_controlled_update
BEFORE UPDATE ON prompt_application_template_drafts
WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id OR NEW.application_id <> OLD.application_id
  OR NEW.owner_subject_ref <> OLD.owner_subject_ref OR NEW.template_id <> OLD.template_id
  OR NEW.draft_version <> OLD.draft_version + 1
  OR json_extract(NEW.sanitized_draft_payload, '$.created_at') <> json_extract(OLD.sanitized_draft_payload, '$.created_at')
  OR json_extract(NEW.sanitized_draft_payload, '$.created_by_actor_ref') <> json_extract(OLD.sanitized_draft_payload, '$.created_by_actor_ref')
BEGIN SELECT RAISE(ABORT, 'prompt application template draft transition is invalid'); END;
CREATE TRIGGER prompt_application_template_drafts_no_delete BEFORE DELETE ON prompt_application_template_drafts
BEGIN SELECT RAISE(ABORT, 'prompt application template drafts cannot be deleted'); END;
CREATE TRIGGER prompt_application_template_versions_no_update BEFORE UPDATE ON prompt_application_template_versions
BEGIN SELECT RAISE(ABORT, 'prompt application template versions are immutable'); END;
CREATE TRIGGER prompt_application_template_versions_no_delete BEFORE DELETE ON prompt_application_template_versions
BEGIN SELECT RAISE(ABORT, 'prompt application template versions cannot be deleted'); END;
