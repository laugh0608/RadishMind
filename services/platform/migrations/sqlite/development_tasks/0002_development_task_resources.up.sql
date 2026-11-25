CREATE TABLE development_task_resources (
 attempt_id TEXT PRIMARY KEY REFERENCES development_tasks(attempt_id),
 owner_instance_id TEXT NOT NULL,
 host_pid INTEGER NOT NULL CHECK(host_pid>0),
 process_namespace TEXT NOT NULL,
 worker_pgid INTEGER NOT NULL DEFAULT 0 CHECK(worker_pgid>=0),
 sdk_worker TEXT NOT NULL DEFAULT 'not_requested' CHECK(sdk_worker IN ('not_requested','pending','active','stopping','confirmed','unconfirmed')),
 code_environment TEXT NOT NULL DEFAULT 'not_requested' CHECK(code_environment IN ('not_requested','pending','active','stopping','confirmed','unconfirmed')),
 model_connection TEXT NOT NULL DEFAULT 'not_requested' CHECK(model_connection IN ('not_requested','active','stopping','confirmed','unconfirmed'))
) STRICT;
