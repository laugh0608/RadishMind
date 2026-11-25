CREATE TABLE development_tasks (
 task_id TEXT PRIMARY KEY, tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL,
 account_id TEXT NOT NULL, application_id TEXT NOT NULL,
 created_at_ms INTEGER NOT NULL CHECK(created_at_ms>0),
 updated_at_ms INTEGER NOT NULL CHECK(updated_at_ms>=created_at_ms),
 record_version INTEGER NOT NULL CHECK(record_version>0),
 plan_revision INTEGER NOT NULL CHECK(plan_revision BETWEEN 1 AND 16),
 attempt_id TEXT UNIQUE, execution_slot INTEGER UNIQUE CHECK(execution_slot IS NULL OR (execution_slot=1 AND attempt_id IS NOT NULL)),
 UNIQUE(tenant_ref,workspace_id,account_id,application_id,task_id),
 FOREIGN KEY(tenant_ref,workspace_id,attempt_id) REFERENCES development_attempts(tenant_ref,workspace_id,attempt_id),
 FOREIGN KEY(task_id,plan_revision) REFERENCES development_task_plans(task_id,plan_revision) DEFERRABLE INITIALLY DEFERRED
) STRICT;
CREATE INDEX development_tasks_history ON development_tasks(tenant_ref,workspace_id,account_id,application_id,created_at_ms,task_id);

CREATE TABLE development_task_plans (
 task_id TEXT NOT NULL REFERENCES development_tasks(task_id), plan_revision INTEGER NOT NULL CHECK(plan_revision BETWEEN 1 AND 16),
 profile_record TEXT NOT NULL CHECK(json_valid(profile_record)), plan_digest TEXT NOT NULL,
 PRIMARY KEY(task_id,plan_revision)
) STRICT;
CREATE TABLE development_task_preparations (
 preparation_id TEXT PRIMARY KEY, task_id TEXT NOT NULL, task_version INTEGER NOT NULL,
 plan_revision INTEGER NOT NULL, prepared_at_ms INTEGER NOT NULL, expires_at_ms INTEGER NOT NULL,
 eligibility_digest TEXT NOT NULL, confirmation_digest TEXT NOT NULL,
 consumed INTEGER NOT NULL DEFAULT 0 CHECK(consumed IN (0,1)),
 CHECK(expires_at_ms>prepared_at_ms AND expires_at_ms-prepared_at_ms<=300000),
 FOREIGN KEY(task_id,plan_revision) REFERENCES development_task_plans(task_id,plan_revision)
) STRICT;
CREATE TABLE development_task_receipts (
 tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, account_id TEXT NOT NULL, application_id TEXT NOT NULL,
 request_id TEXT NOT NULL, operation TEXT NOT NULL CHECK(operation IN ('create_task','revise_plan','prepare','start','cancel','save_artifact')),
 request_digest TEXT NOT NULL, task_id TEXT NOT NULL, plan_revision INTEGER NOT NULL,
 preparation_id TEXT REFERENCES development_task_preparations(preparation_id), attempt_id TEXT, artifact_id TEXT,
 PRIMARY KEY(tenant_ref,workspace_id,account_id,application_id,request_id),
 FOREIGN KEY(tenant_ref,workspace_id,account_id,application_id,task_id)
 REFERENCES development_tasks(tenant_ref,workspace_id,account_id,application_id,task_id),
 FOREIGN KEY(task_id,plan_revision) REFERENCES development_task_plans(task_id,plan_revision),
 FOREIGN KEY(tenant_ref,workspace_id,attempt_id) REFERENCES development_attempts(tenant_ref,workspace_id,attempt_id),
 FOREIGN KEY(tenant_ref,workspace_id,account_id,application_id,artifact_id) REFERENCES development_artifacts(tenant_ref,workspace_id,account_id,application_id,artifact_id),
 CHECK((operation='save_artifact')=(artifact_id IS NOT NULL)),
 CHECK((operation='prepare')=(preparation_id IS NOT NULL)),
 CHECK((operation IN ('start','cancel','save_artifact'))=(attempt_id IS NOT NULL))
) STRICT;
CREATE INDEX development_task_receipts_per_task ON development_task_receipts(task_id);
CREATE TRIGGER development_task_plans_no_update BEFORE UPDATE ON development_task_plans
BEGIN SELECT RAISE(ABORT,'immutable task plan'); END;
CREATE TRIGGER development_task_plans_no_delete BEFORE DELETE ON development_task_plans
BEGIN SELECT RAISE(ABORT,'immutable task plan'); END;
CREATE TRIGGER development_task_receipts_no_update BEFORE UPDATE ON development_task_receipts
BEGIN SELECT RAISE(ABORT,'immutable task receipt'); END;
CREATE TRIGGER development_task_receipts_no_delete BEFORE DELETE ON development_task_receipts
BEGIN SELECT RAISE(ABORT,'immutable task receipt'); END;
