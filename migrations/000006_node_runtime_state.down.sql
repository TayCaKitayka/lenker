ALTER TABLE nodes
    DROP CONSTRAINT IF EXISTS nodes_last_runtime_prepared_revision_check,
    DROP CONSTRAINT IF EXISTS nodes_last_runtime_attempt_status_check,
    DROP CONSTRAINT IF EXISTS nodes_last_dry_run_status_check,
    DROP CONSTRAINT IF EXISTS nodes_runtime_state_check,
    DROP CONSTRAINT IF EXISTS nodes_runtime_mode_check,
    DROP COLUMN IF EXISTS last_runtime_error,
    DROP COLUMN IF EXISTS last_runtime_transition_at,
    DROP COLUMN IF EXISTS last_runtime_prepared_revision,
    DROP COLUMN IF EXISTS last_runtime_attempt_status,
    DROP COLUMN IF EXISTS last_dry_run_status,
    DROP COLUMN IF EXISTS runtime_state,
    DROP COLUMN IF EXISTS runtime_desired_state,
    DROP COLUMN IF EXISTS runtime_mode;
