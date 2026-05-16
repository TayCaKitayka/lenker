ALTER TABLE nodes
    DROP CONSTRAINT IF EXISTS nodes_last_applied_revision_check,
    DROP CONSTRAINT IF EXISTS nodes_last_validation_status_check,
    DROP COLUMN IF EXISTS active_config_path,
    DROP COLUMN IF EXISTS last_applied_revision,
    DROP COLUMN IF EXISTS last_validation_at,
    DROP COLUMN IF EXISTS last_validation_error,
    DROP COLUMN IF EXISTS last_validation_status;
