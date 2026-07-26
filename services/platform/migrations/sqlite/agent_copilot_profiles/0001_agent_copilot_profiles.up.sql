CREATE TABLE agent_copilot_profile_drafts (
    tenant_ref TEXT NOT NULL CHECK (length(trim(tenant_ref)) > 0),
    workspace_id TEXT NOT NULL CHECK (length(trim(workspace_id)) > 0),
    application_id TEXT NOT NULL CHECK (application_id GLOB 'app_[a-z2-7]*' AND length(application_id) = 20 AND substr(application_id, 5) NOT GLOB '*[^a-z2-7]*'),
    owner_subject_ref TEXT NOT NULL CHECK (length(trim(owner_subject_ref)) > 0),
    profile_id TEXT NOT NULL CHECK (profile_id GLOB 'acpf_[a-z2-7]*' AND length(profile_id) = 21 AND substr(profile_id, 6) NOT GLOB '*[^a-z2-7]*'),
    draft_version INTEGER NOT NULL CHECK (draft_version > 0),
    profile_digest TEXT NOT NULL CHECK (profile_digest GLOB 'sha256:[a-f0-9]*' AND length(profile_digest) = 71 AND substr(profile_digest, 8) NOT GLOB '*[^a-f0-9]*'),
    policy_digest TEXT NOT NULL CHECK (policy_digest GLOB 'sha256:[a-f0-9]*' AND length(policy_digest) = 71 AND substr(policy_digest, 8) NOT GLOB '*[^a-f0-9]*'),
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano > 0),
    sanitized_draft_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_draft_payload)
        AND json_type(sanitized_draft_payload) = 'object'
        AND json_extract(sanitized_draft_payload, '$.schema_version') = 'agent_copilot_profile_draft.v1'
        AND json_extract(sanitized_draft_payload, '$.tenant_ref') = tenant_ref
        AND json_extract(sanitized_draft_payload, '$.workspace_id') = workspace_id
        AND json_extract(sanitized_draft_payload, '$.application_id') = application_id
        AND json_extract(sanitized_draft_payload, '$.owner_subject_ref') = owner_subject_ref
        AND json_extract(sanitized_draft_payload, '$.profile_id') = profile_id
        AND json_extract(sanitized_draft_payload, '$.draft_version') = draft_version
        AND json_extract(sanitized_draft_payload, '$.profile_digest') = profile_digest
        AND json_extract(sanitized_draft_payload, '$.policy_digest') = policy_digest
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, profile_id)
) STRICT;

CREATE TABLE agent_copilot_profile_versions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    profile_id TEXT NOT NULL,
    profile_version INTEGER NOT NULL CHECK (profile_version > 0),
    source_draft_version INTEGER NOT NULL CHECK (source_draft_version > 0),
    profile_digest TEXT NOT NULL CHECK (profile_digest GLOB 'sha256:[a-f0-9]*' AND length(profile_digest) = 71 AND substr(profile_digest, 8) NOT GLOB '*[^a-f0-9]*'),
    policy_digest TEXT NOT NULL CHECK (policy_digest GLOB 'sha256:[a-f0-9]*' AND length(policy_digest) = 71 AND substr(policy_digest, 8) NOT GLOB '*[^a-f0-9]*'),
    published_at_unix_nano INTEGER NOT NULL CHECK (published_at_unix_nano > 0),
    immutable_version_payload TEXT NOT NULL CHECK (
        json_valid(immutable_version_payload)
        AND json_type(immutable_version_payload) = 'object'
        AND json_extract(immutable_version_payload, '$.schema_version') = 'agent_copilot_profile_version.v1'
        AND json_extract(immutable_version_payload, '$.tenant_ref') = tenant_ref
        AND json_extract(immutable_version_payload, '$.workspace_id') = workspace_id
        AND json_extract(immutable_version_payload, '$.application_id') = application_id
        AND json_extract(immutable_version_payload, '$.owner_subject_ref') = owner_subject_ref
        AND json_extract(immutable_version_payload, '$.profile_id') = profile_id
        AND json_extract(immutable_version_payload, '$.profile_version') = profile_version
        AND json_extract(immutable_version_payload, '$.source_draft_version') = source_draft_version
        AND json_extract(immutable_version_payload, '$.profile_digest') = profile_digest
        AND json_extract(immutable_version_payload, '$.policy_digest') = policy_digest
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, profile_id, profile_version),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, profile_id, source_draft_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, profile_id)
        REFERENCES agent_copilot_profile_drafts (tenant_ref, workspace_id, application_id, owner_subject_ref, profile_id)
        ON DELETE RESTRICT
) STRICT;

CREATE INDEX agent_copilot_profile_drafts_scope_idx ON agent_copilot_profile_drafts
    (tenant_ref, workspace_id, application_id, owner_subject_ref, updated_at_unix_nano DESC, profile_id ASC);
CREATE INDEX agent_copilot_profile_versions_scope_idx ON agent_copilot_profile_versions
    (tenant_ref, workspace_id, application_id, owner_subject_ref, profile_id, profile_version DESC);

CREATE TRIGGER agent_copilot_profile_drafts_controlled_update
BEFORE UPDATE ON agent_copilot_profile_drafts
WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id OR NEW.application_id <> OLD.application_id
  OR NEW.owner_subject_ref <> OLD.owner_subject_ref OR NEW.profile_id <> OLD.profile_id
  OR NEW.draft_version <> OLD.draft_version + 1
  OR json_extract(NEW.sanitized_draft_payload, '$.created_at') <> json_extract(OLD.sanitized_draft_payload, '$.created_at')
  OR json_extract(NEW.sanitized_draft_payload, '$.created_by_actor_ref') <> json_extract(OLD.sanitized_draft_payload, '$.created_by_actor_ref')
BEGIN SELECT RAISE(ABORT, 'agent copilot profile draft transition is invalid'); END;
CREATE TRIGGER agent_copilot_profile_drafts_no_delete BEFORE DELETE ON agent_copilot_profile_drafts
BEGIN SELECT RAISE(ABORT, 'agent copilot profile drafts cannot be deleted'); END;
CREATE TRIGGER agent_copilot_profile_versions_no_update BEFORE UPDATE ON agent_copilot_profile_versions
BEGIN SELECT RAISE(ABORT, 'agent copilot profile versions are immutable'); END;
CREATE TRIGGER agent_copilot_profile_versions_no_delete BEFORE DELETE ON agent_copilot_profile_versions
BEGIN SELECT RAISE(ABORT, 'agent copilot profile versions cannot be deleted'); END;
