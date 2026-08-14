DROP TRIGGER admin_provider_route_drafts_controlled_update;
DROP TRIGGER admin_provider_route_drafts_no_delete;
DROP TRIGGER admin_provider_route_candidates_controlled_update;
DROP TRIGGER admin_provider_route_candidates_no_delete;
DROP TRIGGER admin_provider_route_reviews_no_update;
DROP TRIGGER admin_provider_route_reviews_no_delete;
DROP TRIGGER admin_provider_route_snapshots_controlled_update;
DROP TRIGGER admin_provider_route_snapshots_no_delete;
DROP TRIGGER admin_provider_route_activations_no_update;
DROP TRIGGER admin_provider_route_activations_no_delete;

ALTER TABLE admin_provider_route_drafts RENAME TO admin_provider_route_drafts_v1;
ALTER TABLE admin_provider_route_candidates RENAME TO admin_provider_route_candidates_v1;
ALTER TABLE admin_provider_route_reviews RENAME TO admin_provider_route_reviews_v1;
ALTER TABLE admin_provider_route_active_snapshots RENAME TO admin_provider_route_active_snapshots_v1;
ALTER TABLE admin_provider_route_activation_records RENAME TO admin_provider_route_activation_records_v1;

CREATE TABLE admin_provider_route_drafts (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development', 'test')),
    configuration_id TEXT NOT NULL,
    draft_revision INTEGER NOT NULL CHECK (draft_revision > 0),
    draft_digest TEXT NOT NULL,
    sanitized_draft_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_draft_payload)
        AND json_extract(sanitized_draft_payload, '$.schema_version') IN (
            'admin_provider_route_configuration_draft.v1',
            'admin_provider_route_configuration_draft.v2'
        )
        AND json_extract(sanitized_draft_payload, '$.tenant_ref') = tenant_ref
        AND json_extract(sanitized_draft_payload, '$.workspace_id') = workspace_id
        AND json_extract(sanitized_draft_payload, '$.environment') = environment
        AND json_extract(sanitized_draft_payload, '$.configuration_id') = configuration_id
        AND json_extract(sanitized_draft_payload, '$.draft_revision') = draft_revision
        AND json_extract(sanitized_draft_payload, '$.draft_digest') = draft_digest
    ),
    updated_at TEXT NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, configuration_id)
) STRICT;

CREATE TABLE admin_provider_route_candidates (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development', 'test')),
    configuration_id TEXT NOT NULL,
    candidate_id TEXT NOT NULL,
    source_draft_revision INTEGER NOT NULL CHECK (source_draft_revision > 0),
    source_draft_digest TEXT NOT NULL,
    candidate_digest TEXT NOT NULL,
    candidate_state TEXT NOT NULL CHECK (candidate_state IN ('pending_review', 'approved', 'rejected')),
    review_version INTEGER NOT NULL CHECK (review_version >= 0),
    sanitized_candidate_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_candidate_payload)
        AND json_extract(sanitized_candidate_payload, '$.schema_version') IN (
            'admin_provider_route_candidate.v1',
            'admin_provider_route_candidate.v2'
        )
        AND json_extract(sanitized_candidate_payload, '$.tenant_ref') = tenant_ref
        AND json_extract(sanitized_candidate_payload, '$.workspace_id') = workspace_id
        AND json_extract(sanitized_candidate_payload, '$.environment') = environment
        AND json_extract(sanitized_candidate_payload, '$.configuration_id') = configuration_id
        AND json_extract(sanitized_candidate_payload, '$.candidate_id') = candidate_id
        AND json_extract(sanitized_candidate_payload, '$.source_draft_revision') = source_draft_revision
        AND json_extract(sanitized_candidate_payload, '$.source_draft_digest') = source_draft_digest
        AND json_extract(sanitized_candidate_payload, '$.candidate_digest') = candidate_digest
        AND json_extract(sanitized_candidate_payload, '$.candidate_state') = candidate_state
        AND json_extract(sanitized_candidate_payload, '$.review_version') = review_version
    ),
    created_at TEXT NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, configuration_id, candidate_id)
) STRICT;

CREATE TABLE admin_provider_route_reviews (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development', 'test')),
    configuration_id TEXT NOT NULL,
    candidate_id TEXT NOT NULL,
    review_version INTEGER NOT NULL CHECK (review_version > 0),
    decision TEXT NOT NULL CHECK (decision IN ('approve', 'reject')),
    resulting_state TEXT NOT NULL CHECK (resulting_state IN ('approved', 'rejected')),
    sanitized_review_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_review_payload)
        AND json_extract(sanitized_review_payload, '$.schema_version') = 'admin_provider_route_review.v1'
        AND json_extract(sanitized_review_payload, '$.review_version') = review_version
        AND json_extract(sanitized_review_payload, '$.decision') = decision
        AND json_extract(sanitized_review_payload, '$.resulting_state') = resulting_state
    ),
    reviewed_at TEXT NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, configuration_id, candidate_id, review_version),
    FOREIGN KEY (tenant_ref, workspace_id, environment, configuration_id, candidate_id)
        REFERENCES admin_provider_route_candidates (
            tenant_ref, workspace_id, environment, configuration_id, candidate_id
        )
) STRICT;

CREATE TABLE admin_provider_route_active_snapshots (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development', 'test')),
    configuration_id TEXT NOT NULL,
    generation INTEGER NOT NULL CHECK (generation > 0),
    candidate_id TEXT NOT NULL,
    candidate_digest TEXT NOT NULL,
    snapshot_digest TEXT NOT NULL,
    sanitized_snapshot_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_snapshot_payload)
        AND json_extract(sanitized_snapshot_payload, '$.schema_version') IN (
            'admin_provider_route_snapshot.v1',
            'admin_provider_route_snapshot.v2'
        )
        AND json_extract(sanitized_snapshot_payload, '$.tenant_ref') = tenant_ref
        AND json_extract(sanitized_snapshot_payload, '$.workspace_id') = workspace_id
        AND json_extract(sanitized_snapshot_payload, '$.environment') = environment
        AND json_extract(sanitized_snapshot_payload, '$.configuration_id') = configuration_id
        AND json_extract(sanitized_snapshot_payload, '$.generation') = generation
        AND json_extract(sanitized_snapshot_payload, '$.candidate_id') = candidate_id
        AND json_extract(sanitized_snapshot_payload, '$.candidate_digest') = candidate_digest
        AND json_extract(sanitized_snapshot_payload, '$.snapshot_digest') = snapshot_digest
    ),
    activated_at TEXT NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, configuration_id),
    FOREIGN KEY (tenant_ref, workspace_id, environment, configuration_id, candidate_id)
        REFERENCES admin_provider_route_candidates (
            tenant_ref, workspace_id, environment, configuration_id, candidate_id
        )
) STRICT;

CREATE TABLE admin_provider_route_activation_records (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development', 'test')),
    configuration_id TEXT NOT NULL,
    after_generation INTEGER NOT NULL CHECK (after_generation > 0),
    activation_id TEXT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('activate', 'rollback')),
    after_candidate_id TEXT NOT NULL,
    after_snapshot_digest TEXT NOT NULL,
    previous_record_digest TEXT NOT NULL,
    record_digest TEXT NOT NULL,
    sanitized_activation_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_activation_payload)
        AND json_extract(sanitized_activation_payload, '$.schema_version') = 'admin_provider_route_activation_record.v1'
        AND json_extract(sanitized_activation_payload, '$.activation_id') = activation_id
        AND json_extract(sanitized_activation_payload, '$.configuration_id') = configuration_id
        AND json_extract(sanitized_activation_payload, '$.action') = action
        AND json_extract(sanitized_activation_payload, '$.after_generation') = after_generation
        AND json_extract(sanitized_activation_payload, '$.after_candidate_id') = after_candidate_id
        AND json_extract(sanitized_activation_payload, '$.after_snapshot_digest') = after_snapshot_digest
        AND json_extract(sanitized_activation_payload, '$.previous_record_digest') = previous_record_digest
        AND json_extract(sanitized_activation_payload, '$.record_digest') = record_digest
    ),
    created_at TEXT NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, configuration_id, after_generation),
    UNIQUE (tenant_ref, workspace_id, environment, configuration_id, activation_id),
    FOREIGN KEY (tenant_ref, workspace_id, environment, configuration_id, after_candidate_id)
        REFERENCES admin_provider_route_candidates (
            tenant_ref, workspace_id, environment, configuration_id, candidate_id
        )
) STRICT;

INSERT INTO admin_provider_route_drafts SELECT * FROM admin_provider_route_drafts_v1;
INSERT INTO admin_provider_route_candidates SELECT * FROM admin_provider_route_candidates_v1;
INSERT INTO admin_provider_route_reviews SELECT * FROM admin_provider_route_reviews_v1;
INSERT INTO admin_provider_route_active_snapshots SELECT * FROM admin_provider_route_active_snapshots_v1;
INSERT INTO admin_provider_route_activation_records SELECT * FROM admin_provider_route_activation_records_v1;

DROP TABLE admin_provider_route_activation_records_v1;
DROP TABLE admin_provider_route_active_snapshots_v1;
DROP TABLE admin_provider_route_reviews_v1;
DROP TABLE admin_provider_route_candidates_v1;
DROP TABLE admin_provider_route_drafts_v1;

CREATE INDEX admin_provider_route_candidates_scope_idx
    ON admin_provider_route_candidates (
        tenant_ref, workspace_id, environment, configuration_id, created_at, candidate_id
    );
CREATE INDEX admin_provider_route_activation_records_scope_idx
    ON admin_provider_route_activation_records (
        tenant_ref, workspace_id, environment, configuration_id, after_generation
    );

CREATE TRIGGER admin_provider_route_drafts_controlled_update
BEFORE UPDATE ON admin_provider_route_drafts
WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id
  OR NEW.environment <> OLD.environment OR NEW.configuration_id <> OLD.configuration_id
  OR NEW.draft_revision <> OLD.draft_revision + 1
  OR json_extract(NEW.sanitized_draft_payload, '$.created_at') <> json_extract(OLD.sanitized_draft_payload, '$.created_at')
  OR json_extract(NEW.sanitized_draft_payload, '$.created_by_actor_ref') <> json_extract(OLD.sanitized_draft_payload, '$.created_by_actor_ref')
BEGIN SELECT RAISE(ABORT, 'admin provider route draft transition is invalid'); END;
CREATE TRIGGER admin_provider_route_drafts_no_delete
BEFORE DELETE ON admin_provider_route_drafts
BEGIN SELECT RAISE(ABORT, 'admin provider route drafts cannot be deleted'); END;

CREATE TRIGGER admin_provider_route_candidates_controlled_update
BEFORE UPDATE ON admin_provider_route_candidates
WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id
  OR NEW.environment <> OLD.environment OR NEW.configuration_id <> OLD.configuration_id
  OR NEW.candidate_id <> OLD.candidate_id OR NEW.source_draft_revision <> OLD.source_draft_revision
  OR NEW.source_draft_digest <> OLD.source_draft_digest OR NEW.candidate_digest <> OLD.candidate_digest
  OR OLD.candidate_state <> 'pending_review' OR NEW.candidate_state NOT IN ('approved', 'rejected')
  OR NEW.review_version <> OLD.review_version + 1 OR NEW.created_at <> OLD.created_at
BEGIN SELECT RAISE(ABORT, 'admin provider route candidate transition is invalid'); END;
CREATE TRIGGER admin_provider_route_candidates_no_delete
BEFORE DELETE ON admin_provider_route_candidates
BEGIN SELECT RAISE(ABORT, 'admin provider route candidates cannot be deleted'); END;

CREATE TRIGGER admin_provider_route_reviews_no_update
BEFORE UPDATE ON admin_provider_route_reviews
BEGIN SELECT RAISE(ABORT, 'admin provider route reviews are immutable'); END;
CREATE TRIGGER admin_provider_route_reviews_no_delete
BEFORE DELETE ON admin_provider_route_reviews
BEGIN SELECT RAISE(ABORT, 'admin provider route reviews cannot be deleted'); END;

CREATE TRIGGER admin_provider_route_snapshots_controlled_update
BEFORE UPDATE ON admin_provider_route_active_snapshots
WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id
  OR NEW.environment <> OLD.environment OR NEW.configuration_id <> OLD.configuration_id
  OR NEW.generation <> OLD.generation + 1
BEGIN SELECT RAISE(ABORT, 'admin provider route snapshot transition is invalid'); END;
CREATE TRIGGER admin_provider_route_snapshots_no_delete
BEFORE DELETE ON admin_provider_route_active_snapshots
BEGIN SELECT RAISE(ABORT, 'admin provider route snapshots cannot be deleted'); END;

CREATE TRIGGER admin_provider_route_activations_no_update
BEFORE UPDATE ON admin_provider_route_activation_records
BEGIN SELECT RAISE(ABORT, 'admin provider route activation records are immutable'); END;
CREATE TRIGGER admin_provider_route_activations_no_delete
BEFORE DELETE ON admin_provider_route_activation_records
BEGIN SELECT RAISE(ABORT, 'admin provider route activation records cannot be deleted'); END;
