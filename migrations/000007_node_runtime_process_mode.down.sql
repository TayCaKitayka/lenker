ALTER TABLE nodes
    DROP CONSTRAINT IF EXISTS nodes_runtime_process_state_check,
    DROP CONSTRAINT IF EXISTS nodes_runtime_process_mode_check,
    DROP COLUMN IF EXISTS runtime_process_state,
    DROP COLUMN IF EXISTS runtime_process_mode;
