CREATE INDEX local_web_sessions_self_service_list_idx
    ON local_web_sessions(user_id, created_at_unix_nano DESC, session_id DESC);
