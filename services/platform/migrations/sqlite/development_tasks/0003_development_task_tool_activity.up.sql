CREATE TABLE development_task_tool_activity (
 attempt_id TEXT PRIMARY KEY REFERENCES development_task_resources(attempt_id),
 request_id TEXT NOT NULL, sequence INTEGER NOT NULL CHECK(sequence BETWEEN 1 AND 3),
 tool_name TEXT NOT NULL CHECK(tool_name IN ('read_file','replace_file','run_tests')),
 candidate_digest TEXT NOT NULL, started_at_ms INTEGER NOT NULL CHECK(started_at_ms>0)
) STRICT;
