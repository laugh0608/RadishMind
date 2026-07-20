CREATE TABLE control_plane_read_schema_versions (
    component text PRIMARY KEY,
    migration_id text NOT NULL,
    store_schema_version text NOT NULL,
    migration_checksum text NOT NULL,
    applied_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE control_plane_tenant_summary_projections (
    tenant_ref text PRIMARY KEY,
    schema_version text NOT NULL,
    projection_version bigint NOT NULL CHECK (projection_version > 0),
    tenant_display_name text NOT NULL,
    tenant_state text NOT NULL,
    plan_ref text NOT NULL,
    quota_summary_ref text NOT NULL,
    deployment_status_ref text NOT NULL,
    audit_summary_ref text NOT NULL,
    projected_at timestamptz NOT NULL
);

CREATE TABLE control_plane_audit_summary_projections (
    tenant_ref text NOT NULL,
    audit_ref text NOT NULL,
    schema_version text NOT NULL,
    projection_version bigint NOT NULL CHECK (projection_version > 0),
    actor_subject_ref text NOT NULL,
    event_kind text NOT NULL,
    resource_ref text NOT NULL,
    decision text NOT NULL,
    failure_code text,
    trace_id text NOT NULL,
    recorded_at timestamptz NOT NULL,
    projected_at timestamptz NOT NULL,
    PRIMARY KEY (tenant_ref, audit_ref)
);

CREATE INDEX control_plane_audit_summary_keyset_idx
    ON control_plane_audit_summary_projections (tenant_ref, recorded_at DESC, audit_ref DESC);

CREATE INDEX control_plane_audit_summary_event_kind_idx
    ON control_plane_audit_summary_projections (tenant_ref, event_kind, recorded_at DESC, audit_ref DESC);

CREATE INDEX control_plane_audit_summary_resource_ref_idx
    ON control_plane_audit_summary_projections (tenant_ref, resource_ref, recorded_at DESC, audit_ref DESC);
